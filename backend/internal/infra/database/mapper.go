package database

/*
Consente di usare la funzione mapper su tutti gli elementi di entityList (di tipo EntityT) per trasformarli
in una lista di elementi di dominio (di tipo DomainT). Se c'è un errore nel mapping ritorna lo slice vuoto (non-nil).

NOTA: il mapper deve restituire anche il tipo error (anche se non necessario), in modo tale da essere compatibile con
i mapper che possono ritornare errori.
*/
func MapEntityListToDomain[EntityT Tabler, DomainT any] (
	entityList []EntityT, mapper func(*EntityT) (DomainT, error),
) (
	domainList []DomainT, err error,
) {
	domainList = make([]DomainT, len(entityList))
	for i, entity := range entityList {
		tenantUser, err := mapper(&entity)
		if err != nil {
			return make([]DomainT, 0), err
		}
		domainList[i] = tenantUser
	}
	return
}