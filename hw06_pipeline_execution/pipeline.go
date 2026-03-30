package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if len(stages) == 0 {
		return in
	}
	if in == nil {
		return nil
	}
	out := make(Bi)
	go func() {
		defer close(out)

		current := in
		for _, stage := range stages {
			current = stage(current)
		}

		for {
			select {
			case <-done:
				return
			case v, ok := <-current:
				if !ok {
					return
				}
				out <- v
			}
		}
	}()

	return out
}
