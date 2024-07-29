}

// MarshalLogObject implements zap.ObjectMarshaler.
func (f ResponseFeatures) MarshalLogObject(objectEncoder zapcore.ObjectEncoder) error {
	objectEncoder.AddBool("acceptResponseError", f.AcceptResponseError)
	return nil
}
