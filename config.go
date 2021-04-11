package rectail

type Config struct {
	StartWith        []string
	RegexpsToWatch   []string
	DelayMillisecond int64
	MaxOffset        int64
	LogPrefix        string
}

func NewDefaultCfg() Config {
	return Config{
		StartWith:        []string{"."},
		RegexpsToWatch:   []string{".*"},
		DelayMillisecond: 600,
		MaxOffset:        400,
		LogPrefix:        "rectail.log",
	}
}
