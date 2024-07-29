			if err := builder.AppendValue(valIdx, values.NewFloat(ts)); err != nil {
				return err
			}
		}

		return nil
	})
}

func (t *timestampTransformation) UpdateWatermark(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateWatermark(pt)
}

func (t *timestampTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *timestampTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
