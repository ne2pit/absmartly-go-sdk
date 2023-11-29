package absmartly

import (
	"context"
	"errors"
	"time"
)

var ErrExpNotFound = errors.New("experiment not found")

type UnitContext struct {
	u  Units
	ab SDK
}

func (uc *UnitContext) GetTreatment(experiment string) (int, error) {
	a, err := uc.GetAssignment(experiment)
	if err != nil {
		return 0, err
	}
	uc.QueueExposure(a)

	return a.Variant(), nil
}

func (uc *UnitContext) PeekTreatment(experiment string) (int, *assignment, error) {
	a, err := uc.GetAssignment(experiment)
	if err != nil {
		return 0, nil, err
	}

	return a.Variant(), a, nil
}

func (uc *UnitContext) GetAssignment(experiment string) (*assignment, error) {
	exp, found := uc.ab.getExperiment(experiment)
	if !found {
		return nil, ErrExpNotFound
	}
	_ = exp.UnitType
	a := &assignment{
		id:   exp.Id,
		name: experiment,
		ts:   time.Now(),
	}

	unitType := exp.UnitType
	unitValue, unitFound := uc.u[unitType]

	switch {
	case exp.Data.FullOnVariant > 0:
		a.variant = exp.Data.FullOnVariant
		a.by = byFullOn
	case unitFound:
		a.variant, a.unitHash = exp.Assigner.Assign(unitValue)
		a.unitType = unitType
		a.by = byNormal
	}

	return a, nil
}

func (uc *UnitContext) QueueExposure(a Assignment) {
	uc.ab.QueueExposure(a)
}

func (uc *UnitContext) PushExposure(ctx context.Context, a *assignment) error {
	return uc.ab.PushExposure(ctx, a)
}

func (uc *UnitContext) Flush(ctx context.Context) error {
	return uc.ab.Flush(ctx)
}
