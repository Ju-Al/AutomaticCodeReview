
  if (_initialized) {
    for (uint8_t n = 0; n <= _aMax; n++) {
      _x += _XA[n];
      _y += _YA[n];
      _z += _ZA[n];
    }

    X = _x / _aMax; // Average available measurements
    Y = _y / _aMax;
    Z = _z / _aMax;

    # if PLUGIN_119_DEBUG

    if (loglevelActiveFor(LOG_LEVEL_DEBUG)) {
      String log;

      if (log.reserve(40)) {
        log  = F("ITG3205: averages, X: ");
        log += X;
        log += F(", Y: ");
        log += Y;
        log += F(", Z: ");
        log += Z;
        addLog(LOG_LEVEL_DEBUG, log);
      }
    }
    # endif // if PLUGIN_119_DEBUG
  }
  return _initialized;
}

// **************************************************************************/
// Initialize ITG3205
// **************************************************************************/
bool P119_data_struct::init_sensor() {
  itg3205 = new (std::nothrow) ITG3205(_i2cAddress);

  if (nullptr != itg3205) {
    addLog(LOG_LEVEL_INFO, F("ITG3205: Initializing Gyro..."));
    itg3205->initGyro();
    addLog(LOG_LEVEL_INFO, F("ITG3205: Calibrating Gyro..."));
    itg3205->calibrate();
    addLog(LOG_LEVEL_INFO, F("ITG3205: Calibration done."));
  } else {
    addLog(LOG_LEVEL_ERROR, F("ITG3205: Initialization of Gyro failed."));
    return false;
  }

  if (loglevelActiveFor(LOG_LEVEL_DEBUG)) {
    String log;

    if (log.reserve(25)) {
      log  = F("ITG3205: Address: 0x");
      log += String(_i2cAddress, HEX);
      addLog(LOG_LEVEL_DEBUG, log);
    }
  }

  return true;
}

#endif // ifdef USES_P119
