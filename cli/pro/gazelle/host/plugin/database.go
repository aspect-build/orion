package plugin

// TODO: move to its own package

type Database struct {
	Symbols []TargetSymbol
}

func (r *Database) AddSymbol(id, provider_type, label, source_path string) {
	r.Symbols = append(r.Symbols, TargetSymbol{
		Symbol: Symbol{
			Id:       id,
			Provider: provider_type,
		},
		Label: label,
	})
	//	fmt.Printf(`
	//
	// Register symbol: %s
	// id: %s
	// label: %s
	// provider_type: %s
	// `, source_path, id, label, provider_type)
}
