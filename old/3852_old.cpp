
  // Enable the magnetometer
  write8(HMC5883_ADDRESS_MAG, HMC5883_REGISTER_MAG_MR_REG_M, 0x00);

  // Set the gain to a known level
  setMagGain(HMC5883_MAGGAIN_1_3);

  return true;
}

/**************************************************************************/
/*!
    @brief  Sets the magnetometer's gain
*/
/**************************************************************************/
void Adafruit_HMC5883_Unified::setMagGain(hmc5883MagGain gain) {
  write8(HMC5883_ADDRESS_MAG, HMC5883_REGISTER_MAG_CRB_REG_M, (byte)gain);

  _magGain = gain;

  switch (gain) {
  case HMC5883_MAGGAIN_1_3:
    _hmc5883_Gauss_LSB_XY = 1100;
    _hmc5883_Gauss_LSB_Z = 980;
    break;
  case HMC5883_MAGGAIN_1_9:
    _hmc5883_Gauss_LSB_XY = 855;
    _hmc5883_Gauss_LSB_Z = 760;
    break;
  case HMC5883_MAGGAIN_2_5:
    _hmc5883_Gauss_LSB_XY = 670;
    _hmc5883_Gauss_LSB_Z = 600;
    break;
  case HMC5883_MAGGAIN_4_0:
    _hmc5883_Gauss_LSB_XY = 450;
    _hmc5883_Gauss_LSB_Z = 400;
    break;
  case HMC5883_MAGGAIN_4_7:
    _hmc5883_Gauss_LSB_XY = 400;
    _hmc5883_Gauss_LSB_Z = 255;
    break;
  case HMC5883_MAGGAIN_5_6:
    _hmc5883_Gauss_LSB_XY = 330;
    _hmc5883_Gauss_LSB_Z = 295;
    break;
  case HMC5883_MAGGAIN_8_1:
    _hmc5883_Gauss_LSB_XY = 230;
    _hmc5883_Gauss_LSB_Z = 205;
    break;
  }
}

/**************************************************************************/
/*!
    @brief  Gets the most recent sensor event
*/
/**************************************************************************/
bool Adafruit_HMC5883_Unified::getEvent(sensors_event_t *event) {
  /* Clear the event */
  memset(event, 0, sizeof(sensors_event_t));

  /* Read new data */
  read();

  event->version = sizeof(sensors_event_t);
  event->sensor_id = _sensorID;
  event->type = SENSOR_TYPE_MAGNETIC_FIELD;
  event->timestamp = 0;
  event->magnetic.x =
      _magData.x / _hmc5883_Gauss_LSB_XY * SENSORS_GAUSS_TO_MICROTESLA;
  event->magnetic.y =
      _magData.y / _hmc5883_Gauss_LSB_XY * SENSORS_GAUSS_TO_MICROTESLA;
  event->magnetic.z =
      _magData.z / _hmc5883_Gauss_LSB_Z * SENSORS_GAUSS_TO_MICROTESLA;

  return true;
}

/**************************************************************************/
/*!
    @brief  Gets the sensor_t data
*/
/**************************************************************************/
void Adafruit_HMC5883_Unified::getSensor(sensor_t *sensor) {
  /* Clear the sensor_t object */
  memset(sensor, 0, sizeof(sensor_t));

  /* Insert the sensor name in the fixed length char array */
  strncpy(sensor->name, "HMC5883", sizeof(sensor->name) - 1);
  sensor->name[sizeof(sensor->name) - 1] = 0;
  sensor->version = 1;
  sensor->sensor_id = _sensorID;
  sensor->type = SENSOR_TYPE_MAGNETIC_FIELD;
  sensor->min_delay = 0;
  sensor->max_value = 800;  // 8 gauss == 800 microTesla
  sensor->min_value = -800; // -8 gauss == -800 microTesla
  sensor->resolution = 0.2; // 2 milligauss == 0.2 microTesla
}
