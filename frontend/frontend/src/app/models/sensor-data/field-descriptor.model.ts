/*
    Mappa la key ad un valore di SensorReading.value
    Rende più leggibile il dato per UI
    Mostra l'unità di misura del dato
*/
export interface FieldDescriptor {
    key: string;
    label: string;
    unit: string;
}