package latex

type Text string
type Command string
type Symbol string

type Verbatim struct {
	Kind string
	Data string
}

type ParameterStart struct {
}

type ParameterEnd struct {
}

type OptionalStart struct {
}

type OptionalEnd struct {
}

type EnvironmentStart struct {
	Name string
}

type EnvironmentEnd struct {
	Name string
}
