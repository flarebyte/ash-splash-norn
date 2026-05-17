// purpose: Defines generation-time data structures used across generation orchestration and code emitters.
// responsibilities: Declare registry/config decode shapes, capability models, and codegen intermediate structures.
// architecture notes: Type centralization avoids drift between orchestration and language-specific emitters.

package engine

type GeneratedArtifact struct {
	Path string
}

type genRegistry struct {
	DesignRegistry struct {
		KeySchemaRegistry map[string]struct {
			SupportedLanguages []string `cue:"supportedLanguages"`
			TranslationPolicy  struct {
				RequireAllSupportedLanguages bool `cue:"requireAllSupportedLanguages"`
			} `cue:"translationPolicy"`
		} `cue:"keySchemaRegistry"`
		GeneratorCapabilities []struct {
			KeySchema         string   `cue:"keySchema"`
			Target            string   `cue:"target"`
			SupportsNodeKinds []string `cue:"supportsNodeKinds"`
			ArtifactPattern   string   `cue:"artifactPattern"`
			ArtifactNaming    struct {
				GoPackageName   string `cue:"goPackageName"`
				DartLibraryName string `cue:"dartLibraryName"`
			} `cue:"artifactNaming"`
		} `cue:"generatorCapabilities"`
	} `cue:"designRegistry"`
}

type genConfig struct {
	I18nEntries []struct {
		Key          string                    `cue:"key"`
		MetaArgs     []string                  `cue:"metaArgs"`
		Description  string                    `cue:"description"`
		Translations map[string]map[string]any `cue:"translations"`
	} `cue:"i18nEntries"`
	TextEntries []struct {
		Key      string   `cue:"key"`
		Value    string   `cue:"value"`
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

type emitDoc map[string]any

type capResolved struct {
	KeySchema       string
	Target          string
	SupportsNode    []string
	ArtifactPattern string
	GoPackageName   string
	DartLibraryName string
}

type commandFlagLit struct {
	Kind    string
	Name    string
	Schema  []string
	Schemas [][]string
}

type commandSpecLit struct {
	CommandPath []string
	AdminOnly   bool
	Flags       []commandFlagLit
}

type codegenData struct {
	keys           []string
	textByKey      map[string]string
	metaByKey      map[string][]string
	validatorByKey map[string]commandSpecLit
}
