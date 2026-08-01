package llm

type StreamChunk struct {
	Text string
	Done bool
	Err  error
}
