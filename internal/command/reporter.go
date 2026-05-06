package command

type NoopReporter struct{}

func (NoopReporter) UpdateStep(StepUpdate) {}
