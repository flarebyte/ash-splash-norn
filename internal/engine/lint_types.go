package engine

type lintRegistry struct {
	DesignRegistry struct {
		KeySchemaRegistry     map[string]keySchemaSpec `cue:"keySchemaRegistry"`
		GeneratorCapabilities []struct {
			Target          string `cue:"target"`
			ArtifactPattern string `cue:"artifactPattern"`
		} `cue:"generatorCapabilities"`
	} `cue:"designRegistry"`
}

type lintConfig struct {
	I18nEntries []struct {
		Key          string                    `cue:"key"`
		Translations map[string]map[string]any `cue:"translations"`
		MetaArgs     []string                  `cue:"metaArgs"`
	} `cue:"i18nEntries"`
	TextEntries []struct {
		Key      string   `cue:"key"`
		MetaArgs []string `cue:"metaArgs"`
	} `cue:"textEntries"`
	Validations []struct {
		Key      string   `cue:"key"`
		MetaArgs []string `cue:"metaArgs"`
		Commands []struct {
			Args map[string]any `cue:"args"`
		} `cue:"commands"`
	} `cue:"validations"`
}

type keySchemaSpec struct {
	Metadata struct {
		ID      string `cue:"id"`
		Version string `cue:"version"`
	} `cue:"metadata"`
	SupportedLanguages       []string `cue:"supportedLanguages"`
	SupportedCommandSections []string `cue:"supportedCommandSections"`
	TranslationPolicy        struct {
		RequireAllSupportedLanguages bool `cue:"requireAllSupportedLanguages"`
	} `cue:"translationPolicy"`
	MetaArgsValidation struct {
		Args map[string]struct {
			CommandPath []string `cue:"commandPath"`
			AdminOnly   bool     `cue:"adminOnly"`
			Flags       []struct {
				Kind    string     `cue:"kind"`
				Name    string     `cue:"name"`
				Schema  []string   `cue:"schema"`
				Schemas [][]string `cue:"schemas"`
			} `cue:"flags"`
		} `cue:"args"`
	} `cue:"metaArgsValidation"`
	RootLabels   []string `cue:"rootLabels"`
	NodesByLabel map[string]struct {
		Label       string   `cue:"label"`
		Kind        string   `cue:"kind"`
		Mandatory   bool     `cue:"mandatory"`
		ChildLabels []string `cue:"childLabels"`
	} `cue:"nodesByLabel"`
}

type expectedByKind struct {
	I18n      map[string]struct{}
	Text      map[string]struct{}
	Validator map[string]struct{}
}

type mandatoryByKind struct {
	I18n      map[string]struct{}
	Text      map[string]struct{}
	Validator map[string]struct{}
}
