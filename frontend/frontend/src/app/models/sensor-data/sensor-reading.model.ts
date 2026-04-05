export interface SensorReading {
  timestamp: string;
  /*
    Per mappare i valori dei sensori in questo modo:
    HEART_RATE: {
      "bpm": 72
    },
    PULSE_OXIMETER: {
      "spo2": 98
      "pulse_rate": 72
    }
    HEALTH_THERMOMETER: {
      "temperature": 36.5
    }
    ENVIROMENTAL_SENSOR: {
      "temperature": 22.5,
      "humidity": 60,
      "pressure": 1013
    }
    ECG_CUSTOM: {
      boh devo ancora capire 
    }
  */
  value: Record<string, number>;
}
