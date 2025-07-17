package watch

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/aspect-build/silo/cli/core/buildinfo"
	cache "github.com/aspect-build/silo/cli/core/gazelle/common/cache"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	watcher "github.com/aspect-build/silo/cli/pro/pkg/watch"
	"github.com/bazelbuild/bazel-gazelle/config"
)

func init() {
	gob.Register(cacheState{})
}

type cacheState struct {
	ClockSpec string
	BuildId   string
	Entries   map[string]any
}

type watchmanCache struct {
	w *watcher.WatchmanWatcher

	file string

	old           map[string]any
	new           *sync.Map
	lastClockSpec string
}

var _ cache.Cache = (*watchmanCache)(nil)

func NewWatchmanCache(c *config.Config) cache.Cache {
	diskCachePath := os.Getenv("ASPECT_CONFIGURE_CACHE")
	if diskCachePath == "" {
		// A default path for the cache file.
		// Try to be unique per repo to allow re-use, while using a temp dir to avoid clutter and indicate
		// the cache is not required.
		diskCachePath = path.Join(os.TempDir(), fmt.Sprintf("aspect-configure-%v.cache", c.RepoName))
	}

	// Start the watcher
	w := watcher.NewWatchman(c.RepoRoot)
	if err := w.Start(); err != nil {
		log.Fatalf("failed to start the watcher: %v", err)
	}

	wc := &watchmanCache{
		w:    w,
		file: diskCachePath,
		old:  map[string]any{},
		new:  &sync.Map{},
	}
	wc.read()

	runtime.SetFinalizer(wc, closeWatchmanCache)

	return wc
}

func closeWatchmanCache(c *watchmanCache) {
	c.w.Close()
}

func (c *watchmanCache) read() {
	cacheReader, err := os.Open(c.file)
	if err != nil {
		BazelLog.Tracef("Failed to open cache %q: %v", c.file, err)
		return
	}
	defer cacheReader.Close()

	var v cacheState

	cacheDecoder := gob.NewDecoder(cacheReader)
	if e := cacheDecoder.Decode(&v); e != nil {
		BazelLog.Errorf("Failed to read cache %q: %v", c.file, e)
		return
	}

	// If the stamp has changed, discard the cache.
	if buildinfo.IsStamped() {
		if v.BuildId != buildinfo.GitCommit {
			BazelLog.Infof("Cache buildId stale, clearing")
			return
		}
	}

	cs, err := c.w.GetDiff(v.ClockSpec)
	if err != nil {
		BazelLog.Errorf("Failed to get diff from watchman: %v", err)
		return
	}

	// If the watcher has restarted, discard the cache.
	if cs.IsFreshInstance {
		BazelLog.Infof("Watchman state is stale, clearing")
		return
	}

	// Discard entries which have changed since the last cache write.
	for _, p := range cs.Paths {
		v.Entries[p] = nil
	}

	// Persist the still valid old entries and latest clock spec.
	for k, v := range v.Entries {
		if v != nil {
			c.old[k] = v
		}
	}
	c.lastClockSpec = cs.ClockSpec

	BazelLog.Infof("Watchman cache: %d entries at clock spec %s", len(c.old), c.lastClockSpec)
}

func (c *watchmanCache) write() {
	cacheWriter, err := os.OpenFile(c.file, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		BazelLog.Errorf("Failed to create cache %q: %v", c.file, err)
		return
	}
	defer cacheWriter.Close()

	m := make(map[string]any)

	// Convert the sync.Map to a regular map for serialization.
	c.new.Range(func(key, value interface{}) bool {
		if m, isMap := value.(*sync.Map); isMap {
			mValue := make(map[string]any)
			m.Range(func(k, v interface{}) bool {
				mValue[k.(string)] = v
				return true
			})
			value = mValue
		}
		m[key.(string)] = value
		return true
	})

	// Include the clock spec and build id in the cache.
	s := cacheState{
		ClockSpec: c.lastClockSpec,
		Entries:   m,
		BuildId:   "",
	}

	if buildinfo.IsStamped() {
		s.BuildId = buildinfo.GitCommit
	}

	cacheEncoder := gob.NewEncoder(cacheWriter)
	if e := cacheEncoder.Encode(s); e != nil {
		BazelLog.Errorf("Failed to write cache %q: %v", c.file, e)
	}
}

func (c *watchmanCache) Load(key string) (any, bool) {
	// Already written to new cache.
	if v, found := c.new.Load(key); found {
		return v, true
	}

	// Exists in old cache and can transfer to new.
	if v, ok := c.old[key]; ok {
		v, _ = c.LoadOrStore(key, v)
		return v, true
	}

	// Uncached
	return nil, false
}

func (c *watchmanCache) Store(key string, value any) {
	c.new.Store(key, value)
}

func (c *watchmanCache) LoadOrStore(key string, value any) (any, bool) {
	return c.new.LoadOrStore(key, value)
}

func (c *watchmanCache) Persist() {
	c.write()
}

func (c *watchmanCache) LoadOrStoreFile(root, p, key string, loader cache.FileCompute) (any, bool, error) {
	// Load directly from c.new to potentially convert map[] to sync.Map
	fileMap, hasFileMap := c.new.Load(p)
	if !hasFileMap {
		// A new map for this file path
		newMap := &sync.Map{}

		// Potentially load the previously persisted map[]
		if oldMap, hasOld := c.old[p]; hasOld {
			for k, v := range oldMap.(map[string]any) {
				newMap.Store(k, v)
			}
		}

		fileMap, hasFileMap = c.LoadOrStore(p, newMap)
	}

	// Load any cached result from the file specific sync.Map
	v, found := fileMap.(*sync.Map).Load(key)
	if found {
		return v, true, nil
	}

	// Uncached and must be computed from file content
	content, err := os.ReadFile(path.Join(root, p))
	if err != nil {
		return nil, false, err
	}

	// Compute the value and store it in the file specific sync.Map
	v, err = loader(p, content)
	if err == nil {
		v, found = fileMap.(*sync.Map).LoadOrStore(key, v)
	}

	return v, found, err
}
