import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WritableSignal, signal } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { By } from '@angular/platform-browser';
import { of, Subject } from 'rxjs';
import { describe, expect, it, vi, afterEach } from 'vitest';

import { GatewaySensorManagerPage } from './gateway-sensor-manager.page';
import { GatewayTableComponent } from '../shared/components/gateway-table/gateway-table.component';
import { GatewayExpandedComponent } from '../shared/components/gateway-expanded/gateway-expanded.component';
import { SensorTableComponent } from '../shared/components/sensor-table/sensor-table.component';
import { GatewaySensorManagerService } from '../../services/gateway-sensor-manager/gateway-sensor-manager.service';
import { ConfirmDeleteDialog } from '../shared/dialogs/confirm-delete/confirm-delete.dialog';
import { CreateGatewayDialog } from './dialogs/create-gateway/create-gateway.dialog';
import { CreateSensorDialog } from './dialogs/create-sensor/create-sensor.dialog';
import { Gateway } from '../../models/gateway/gateway.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorStatus } from '../../models/sensor-status.enum';
import { GatewayStatus } from '../../models/gateway-status.enum';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';

const mockGateways: Gateway[] = [
  {
    id: 'gw-1',
    tenantId: 'tenant-1',
    name: 'Gateway Alpha',
    status: GatewayStatus.ACTIVE,
    interval: 60,
  },
  {
    id: 'gw-2',
    tenantId: 'tenant-1',
    name: 'Gateway Beta',
    status: GatewayStatus.INACTIVE,
    interval: 120,
  },
];

const mockSensors: Sensor[] = [
  {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Temperature',
    profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    status: SensorStatus.ACTIVE,
    dataInterval: 30,
  },
  {
    id: 'sensor-2',
    gatewayId: 'gw-1',
    name: 'Humidity',
    profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
    status: SensorStatus.ACTIVE,
    dataInterval: 30,
  },
];

function createManagerServiceMock() {
  return {
    gatewayList: signal<Gateway[]>([]),
    gatewayTotal: signal(0),
    gatewayPageIndex: signal(0),
    gatewayLimit: signal(10),
    gatewayLoading: signal(false),
    gatewayError: signal<string | null>(null),

    sensorList: signal<Sensor[]>([]),
    sensorTotal: signal(0),
    sensorPageIndex: signal(0),
    sensorLimit: signal(10),
    sensorLoading: signal(false),
    sensorError: signal<string | null>(null),

    expandedGateway: signal<Gateway | null>(null),

    loadGateways: vi.fn(),
    toggleExpandedGateway: vi.fn(),
    changeGatewayPage: vi.fn(),
    changeSensorPage: vi.fn(),
    refreshGateways: vi.fn(),
    refreshSensors: vi.fn(),
    deleteGateway: vi.fn().mockReturnValue(of(void 0)),
    deleteSensor: vi.fn().mockReturnValue(of(void 0)),
  };
}

function setupTestBed() {
  const managerServiceMock = createManagerServiceMock();
  const snackBarMock = { open: vi.fn() };

  let activeSubject = new Subject<unknown>();
  const dialogMock = {
    open: vi.fn().mockImplementation(() => ({
      afterClosed: () => activeSubject.asObservable(),
    })),
  };

  const confirmDialog = () => {
    activeSubject.next(true);
    activeSubject.complete();
    activeSubject = new Subject<unknown>();
  };

  const cancelDialog = () => {
    activeSubject.next(false);
    activeSubject.complete();
    activeSubject = new Subject<unknown>();
  };

  TestBed.configureTestingModule({
    imports: [
      GatewaySensorManagerPage,
      GatewayTableComponent,
      GatewayExpandedComponent,
      SensorTableComponent,
    ],
    providers: [{ provide: GatewaySensorManagerService, useValue: managerServiceMock }],
  })
    .overrideProvider(MatDialog, { useValue: dialogMock })
    .overrideProvider(MatSnackBar, { useValue: snackBarMock });

  const fixture = TestBed.createComponent(GatewaySensorManagerPage);
  return { fixture, managerServiceMock, dialogMock, snackBarMock, confirmDialog, cancelDialog };
}

function setupWithExpanded(
  managerServiceMock: ReturnType<typeof createManagerServiceMock>,
  fixture: ComponentFixture<GatewaySensorManagerPage>,
): void {
  (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
  (managerServiceMock.expandedGateway as WritableSignal<Gateway | null>).set(mockGateways[0]);
  (managerServiceMock.sensorList as WritableSignal<Sensor[]>).set(mockSensors);
  fixture.detectChanges();
}

const getGatewayTable = (f: ComponentFixture<GatewaySensorManagerPage>) =>
  f.debugElement.query(By.directive(GatewayTableComponent));
const getExpandedComponent = (f: ComponentFixture<GatewaySensorManagerPage>) =>
  f.debugElement.query(By.directive(GatewayExpandedComponent));
const getSensorTable = (f: ComponentFixture<GatewaySensorManagerPage>) =>
  f.debugElement.query(By.directive(SensorTableComponent));
const getGatewayRows = (f: ComponentFixture<GatewaySensorManagerPage>): HTMLElement[] =>
  Array.from((f.nativeElement as HTMLElement).querySelectorAll('mat-row:not(.detail-row)'));
const getGatewayDeleteButtons = (
  f: ComponentFixture<GatewaySensorManagerPage>,
): HTMLButtonElement[] =>
  Array.from(
    (f.nativeElement as HTMLElement).querySelectorAll<HTMLButtonElement>(
      'mat-row:not(.detail-row) mat-cell button[color="warn"]',
    ),
  );
const getGatewayCreateButton = (
  f: ComponentFixture<GatewaySensorManagerPage>,
): HTMLButtonElement | null =>
  (f.nativeElement as HTMLElement).querySelector<HTMLButtonElement>(
    '.manager-header button[mat-raised-button]',
  );
const getSensorDeleteButtons = (
  f: ComponentFixture<GatewaySensorManagerPage>,
): HTMLButtonElement[] =>
  Array.from(
    (f.nativeElement as HTMLElement).querySelectorAll<HTMLButtonElement>(
      '.expanded-content mat-cell button[color="warn"]',
    ),
  );
const getSensorCreateButton = (
  f: ComponentFixture<GatewaySensorManagerPage>,
): HTMLButtonElement | null =>
  (f.nativeElement as HTMLElement).querySelector<HTMLButtonElement>(
    '.expanded-content .manager-header button[mat-raised-button]',
  );

describe('GatewaySensorManagerPage (Integration)', () => {
  afterEach(() => {
    TestBed.resetTestingModule();
    vi.resetAllMocks();
  });

  describe('Initialization', () => {
    it('should call loadGateways and render manage mode UI', () => {
      const { fixture, managerServiceMock } = setupTestBed();
      (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
      fixture.detectChanges();

      expect(managerServiceMock.loadGateways).toHaveBeenCalledOnce();
      const gatewayTable = getGatewayTable(fixture);
      expect(gatewayTable).toBeTruthy();
      expect(gatewayTable.componentInstance.actionMode()).toBe('manage');
      expect(fixture.nativeElement.querySelector('.manager-header h1').textContent).toContain(
        'Gestione Gateway',
      );
      expect(getGatewayCreateButton(fixture)!.textContent).toContain('Nuovo Gateway');
    });
  });

  describe('GatewayTable Input Bindings', () => {
    it('should render gateway rows with delete action column in manage mode', () => {
      const { fixture, managerServiceMock } = setupTestBed();
      (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
      fixture.detectChanges();

      expect(getGatewayRows(fixture).length).toBe(2);
      const headerTexts = Array.from<Element>(
        fixture.nativeElement.querySelectorAll('mat-header-cell'),
      ).map((h) => h.textContent?.trim());
      expect(headerTexts).toContain('Azioni');
    });

    it('should display correct gateway data in cells', () => {
      const { fixture, managerServiceMock } = setupTestBed();
      (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set([mockGateways[0]]);
      fixture.detectChanges();

      const cellTexts = Array.from<Element>(
        fixture.nativeElement.querySelectorAll('mat-row:not(.detail-row) mat-cell'),
      ).map((c) => c.textContent?.trim());
      expect(cellTexts).toContain('Gateway Alpha');
      expect(cellTexts).toContain('tenant-1');
      expect(cellTexts).toContain('ATTIVO');
    });

    it('should show spinner when gateway loading', () => {
      const { fixture, managerServiceMock } = setupTestBed();
      (managerServiceMock.gatewayLoading as WritableSignal<boolean>).set(true);
      fixture.detectChanges();

      expect(fixture.nativeElement.querySelector('mat-spinner')).toBeTruthy();
    });

    it('should show empty state when no gateways', () => {
      const { fixture, managerServiceMock } = setupTestBed();
      (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set([]);
      fixture.detectChanges();

      const emptyState = fixture.nativeElement.querySelector('.empty-state');
      expect(emptyState).toBeTruthy();
      expect(emptyState.textContent).toContain('Nessun gateway disponibile');
    });

    it.each([
      { label: 'gateway', signalKey: 'gatewayError' as const, message: 'Load failed' },
      { label: 'sensor', signalKey: 'sensorError' as const, message: 'Sensor load failed' },
    ])('should show error banner for $label error', ({ signalKey, message }) => {
      const { fixture, managerServiceMock } = setupTestBed();
      (managerServiceMock[signalKey] as WritableSignal<string | null>).set(message);
      fixture.detectChanges();

      const errorBanner = fixture.nativeElement.querySelector('.error-banner');
      expect(errorBanner).toBeTruthy();
      expect(errorBanner.textContent).toContain(message);
    });

    describe('Gateway Expand -> Sensor Table', () => {
      it('should NOT show expanded component when no gateway is expanded', () => {
        const { fixture, managerServiceMock } = setupTestBed();
        (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
        fixture.detectChanges();

        expect(getExpandedComponent(fixture)).toBeFalsy();
      });

      it('should render full expanded view: sensors, heading, manage header, and delete buttons', () => {
        const { fixture, managerServiceMock } = setupTestBed();
        setupWithExpanded(managerServiceMock, fixture);

        expect(getExpandedComponent(fixture)).toBeTruthy();
        expect(getSensorTable(fixture)).toBeTruthy();

        const cellTexts = Array.from<Element>(
          fixture.nativeElement.querySelectorAll('.expanded-content mat-row mat-cell'),
        ).map((c) => c.textContent?.trim());
        expect(cellTexts).toContain('Temperature');
        expect(cellTexts).toContain('Humidity');

        expect(fixture.nativeElement.querySelector('.expanded-content h2').textContent).toContain(
          'Gateway Alpha',
        );
        expect(
          fixture.nativeElement.querySelector('.expanded-content .manager-header h3').textContent,
        ).toContain('Gestione Sensori');
        expect(getSensorCreateButton(fixture)!.textContent).toContain('Nuovo Sensore');
        expect(getSensorDeleteButtons(fixture).length).toBe(2);
      });

      it('should show sensor spinner when sensor loading in expanded view', () => {
        const { fixture, managerServiceMock } = setupTestBed();
        (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
        (managerServiceMock.expandedGateway as WritableSignal<Gateway | null>).set(mockGateways[0]);
        (managerServiceMock.sensorLoading as WritableSignal<boolean>).set(true);
        fixture.detectChanges();

        expect(fixture.nativeElement.querySelector('.expanded-content mat-spinner')).toBeTruthy();
      });
    });

    describe('Event Forwarding', () => {
      it('should call toggleExpandedGateway when gateway row is clicked', () => {
        const { fixture, managerServiceMock } = setupTestBed();
        (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
        fixture.detectChanges();

        getGatewayRows(fixture)[0].click();

        expect(managerServiceMock.toggleExpandedGateway).toHaveBeenCalledWith(mockGateways[0]);
      });

      it('should call changeGatewayPage when gateway paginator emits', () => {
        const { fixture, managerServiceMock } = setupTestBed();
        (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
        fixture.detectChanges();

        getGatewayTable(fixture).componentInstance.gatewayPageChange.emit({
          pageIndex: 2,
          pageSize: 25,
          length: 50,
        });

        expect(managerServiceMock.changeGatewayPage).toHaveBeenCalledWith(2, 25);
      });

      it('should call changeSensorPage when sensor paginator emits', () => {
        const { fixture, managerServiceMock } = setupTestBed();
        setupWithExpanded(managerServiceMock, fixture);

        getSensorTable(fixture).componentInstance.pageChange.emit({
          pageIndex: 1,
          pageSize: 10,
          length: 20,
        });

        expect(managerServiceMock.changeSensorPage).toHaveBeenCalledWith(1, 10);
      });
    });

    describe('Gateway Create Flow', () => {
      it('should open CreateGatewayDialog and refresh on success', () => {
        const { fixture, managerServiceMock, dialogMock, snackBarMock, confirmDialog } =
          setupTestBed();
        (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
        fixture.detectChanges();

        getGatewayCreateButton(fixture)!.click();
        fixture.detectChanges();

        expect(dialogMock.open).toHaveBeenCalledWith(CreateGatewayDialog);

        confirmDialog();

        expect(managerServiceMock.refreshGateways).toHaveBeenCalled();
        expect(snackBarMock.open).toHaveBeenCalledWith('Gateway creato con successo', 'Chiudi', {
          duration: 3000,
        });
      });

      it('should NOT refresh when dialog is cancelled', () => {
        const { fixture, managerServiceMock, cancelDialog } = setupTestBed();
        (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
        fixture.detectChanges();

        getGatewayCreateButton(fixture)!.click();
        cancelDialog();

        expect(managerServiceMock.refreshGateways).not.toHaveBeenCalled();
      });
    });

    describe('Gateway Delete Flow', () => {
      it.each([
        { idx: 0, gateway: mockGateways[0], name: 'Gateway Alpha' },
        { idx: 1, gateway: mockGateways[1], name: 'Gateway Beta' },
      ])(
        'should open ConfirmDeleteDialog for "$name" and delete on confirm',
        ({ idx, gateway }) => {
          const { fixture, managerServiceMock, dialogMock, snackBarMock, confirmDialog } =
            setupTestBed();
          (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
          fixture.detectChanges();

          getGatewayDeleteButtons(fixture)[idx].click();
          fixture.detectChanges();

          expect(dialogMock.open).toHaveBeenCalledWith(ConfirmDeleteDialog, {
            data: {
              title: 'Elimina Gateway',
              message: `Sei sicuro di voler eliminare il gateway "${gateway.name}"?`,
            },
          });

          confirmDialog();

          expect(managerServiceMock.deleteGateway).toHaveBeenCalledWith(gateway);
          expect(snackBarMock.open).toHaveBeenCalledWith(
            'Gateway eliminato con successo',
            'Chiudi',
            {
              duration: 3000,
            },
          );
        },
      );

      it('should NOT delete when dialog is cancelled', () => {
        const { fixture, managerServiceMock, cancelDialog } = setupTestBed();
        (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set(mockGateways);
        fixture.detectChanges();

        getGatewayDeleteButtons(fixture)[0].click();
        cancelDialog();

        expect(managerServiceMock.deleteGateway).not.toHaveBeenCalled();
      });
    });

    describe('Sensor Create Flow', () => {
      it('should open CreateSensorDialog with gateway data and refresh on success', () => {
        const { fixture, managerServiceMock, dialogMock, snackBarMock, confirmDialog } =
          setupTestBed();
        setupWithExpanded(managerServiceMock, fixture);

        getSensorCreateButton(fixture)!.click();
        fixture.detectChanges();

        expect(dialogMock.open).toHaveBeenCalledWith(CreateSensorDialog, {
          data: { id: 'gw-1', name: 'Gateway Alpha' },
        });

        confirmDialog();

        expect(managerServiceMock.refreshSensors).toHaveBeenCalledWith('gw-1');
        expect(snackBarMock.open).toHaveBeenCalledWith('Sensore creato con successo', 'Chiudi', {
          duration: 3000,
        });
      });

      it('should NOT refresh when dialog is cancelled', () => {
        const { fixture, managerServiceMock, cancelDialog } = setupTestBed();
        setupWithExpanded(managerServiceMock, fixture);

        getSensorCreateButton(fixture)!.click();
        cancelDialog();

        expect(managerServiceMock.refreshSensors).not.toHaveBeenCalled();
      });
    });

    describe('Sensor Delete Flow', () => {
      it.each([
        { idx: 0, sensor: mockSensors[0], name: 'Temperature' },
        { idx: 1, sensor: mockSensors[1], name: 'Humidity' },
      ])('should open ConfirmDeleteDialog for "$name" and delete on confirm', ({ idx, sensor }) => {
        const { fixture, managerServiceMock, dialogMock, snackBarMock, confirmDialog } =
          setupTestBed();
        setupWithExpanded(managerServiceMock, fixture);

        getSensorDeleteButtons(fixture)[idx].click();
        fixture.detectChanges();

        expect(dialogMock.open).toHaveBeenCalledWith(ConfirmDeleteDialog, {
          data: {
            title: 'Elimina Sensore',
            message: `Sei sicuro di voler eliminare il sensore "${sensor.name}"?`,
          },
        });

        confirmDialog();

        expect(managerServiceMock.deleteSensor).toHaveBeenCalledWith(sensor);
        expect(snackBarMock.open).toHaveBeenCalledWith('Sensore eliminato con successo', 'Chiudi', {
          duration: 3000,
        });
      });

      it('should NOT delete when dialog is cancelled', () => {
        const { fixture, managerServiceMock, cancelDialog } = setupTestBed();
        setupWithExpanded(managerServiceMock, fixture);

        getSensorDeleteButtons(fixture)[0].click();
        cancelDialog();

        expect(managerServiceMock.deleteSensor).not.toHaveBeenCalled();
      });
    });
  });
});
