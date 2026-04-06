import { describe, it, expect, vi, beforeEach } from 'vitest';
import { signal, WritableSignal, Component, input, output } from '@angular/core';
import { TestBed, ComponentFixture } from '@angular/core/testing';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { PageEvent } from '@angular/material/paginator';
import { By } from '@angular/platform-browser';
import { of } from 'rxjs';

import { GatewaySensorManagerPage } from './gateway-sensor-manager.page';
import { GatewayTableComponent } from '../shared/components/gateway-table/gateway-table.component';
import { GatewaySensorManagerService } from '../../services/gateway-sensor-manager/gateway-sensor-manager.service';
import { ConfirmDeleteDialog } from '../shared/dialogs/confirm-delete/confirm-delete.dialog';
import { CreateGatewayDialog } from './dialogs/create-gateway/create-gateway.dialog';
import { CreateSensorDialog } from './dialogs/create-sensor/create-sensor.dialog';
import { Gateway } from '../../models/gateway/gateway.model';
import { Sensor } from '../../models/sensor/sensor.model';
import { Status } from '../../models/gateway-sensor-status.enum';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';
import { ActionMode } from '../../models/action-mode.model';

@Component({ selector: 'app-gateway-table', template: '', standalone: true })
class StubGatewayTable {
  actionMode = input<ActionMode>();
  gateways = input<Gateway[]>();
  sensors = input<Sensor[]>();
  expandedGateway = input<Gateway | null>();
  gatewayLoading = input<boolean>();
  sensorLoading = input<boolean>();
  gatewayTotal = input<number>();
  gatewayPageIndex = input<number>();
  gatewayLimit = input<number>();
  sensorTotal = input<number>();
  sensorPageIndex = input<number>();
  sensorLimit = input<number>();
  gatewayCreateRequested = output<void>();
  gatewayDeleteRequested = output<Gateway>();
  sensorCreateRequested = output<Gateway>();
  sensorDeleteRequested = output<Sensor>();
  expandedGatewayChange = output<Gateway>();
  gatewayPageChange = output<PageEvent>();
  sensorPageChange = output<PageEvent>();
}

describe('GatewaySensorManagerPage (Unit)', () => {
  let fixture: ComponentFixture<GatewaySensorManagerPage>;
  let component: GatewaySensorManagerPage;
  let gatewayErrorSig: WritableSignal<string | null>;
  let sensorErrorSig: WritableSignal<string | null>;

  const mockGateway: Gateway = {
    id: 'gw-1',
    tenantId: 'tenant-01',
    name: 'Gateway 1',
    status: Status.ACTIVE,
    interval: 60,
  };
  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Heart Rate Sensor',
    profile: SensorProfiles.HEART_RATE_SERVICE,
    status: Status.ACTIVE,
    dataInterval: 1000,
  };

  let managerServiceMock: any;
  let dialogMock: { open: ReturnType<typeof vi.fn> };
  let snackBarMock: { open: ReturnType<typeof vi.fn> };

  beforeEach(async () => {
    vi.resetAllMocks();
    gatewayErrorSig = signal<string | null>(null);
    sensorErrorSig = signal<string | null>(null);

    managerServiceMock = {
      gatewayList: signal<Gateway[]>([]),
      gatewayTotal: signal(0),
      gatewayPageIndex: signal(0),
      gatewayLimit: signal(10),
      gatewayLoading: signal(false),
      gatewayError: gatewayErrorSig,
      sensorList: signal<Sensor[]>([]),
      sensorTotal: signal(0),
      sensorPageIndex: signal(0),
      sensorLimit: signal(10),
      sensorLoading: signal(false),
      sensorError: sensorErrorSig,
      expandedGateway: signal<Gateway | null>(null),
      loadGateways: vi.fn(),
      toggleExpandedGateway: vi.fn(),
      changeGatewayPage: vi.fn(),
      changeSensorPage: vi.fn(),
      refreshGateways: vi.fn(),
      refreshSensors: vi.fn(),
      deleteGateway: vi.fn().mockReturnValue(of(undefined)),
      deleteSensor: vi.fn().mockReturnValue(of(undefined)),
    };

    dialogMock = { open: vi.fn() };
    snackBarMock = { open: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [GatewaySensorManagerPage],
      providers: [
        { provide: GatewaySensorManagerService, useValue: managerServiceMock },
        { provide: MatDialog, useValue: dialogMock },
        { provide: MatSnackBar, useValue: snackBarMock },
      ],
    })
      .overrideComponent(GatewaySensorManagerPage, {
        remove: { imports: [GatewayTableComponent] },
        add: { imports: [StubGatewayTable] },
      })
      .compileComponents();

    fixture = TestBed.createComponent(GatewaySensorManagerPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  const getTable = () => fixture.debugElement.query(By.directive(StubGatewayTable));
  const mockDialog = (result: any) => {
    dialogMock.open.mockReturnValue({ afterClosed: () => of(result) });
  };

  describe('ngOnInit', () => {
    it('should call loadGateways on init', () => {
      expect(managerServiceMock.loadGateways).toHaveBeenCalledOnce();
    });
  });

  describe('error banner', () => {
    it.each([
      ['Gateway failed', null, 'Gateway failed'],
      [null, 'Sensor failed', 'Sensor failed'],
      ['GW error', 'Sens error', 'GW error'],
    ])('should show correct error (gateway=%s, sensor=%s)', (gwErr, sensErr, expected) => {
      gatewayErrorSig.set(gwErr);
      sensorErrorSig.set(sensErr);
      fixture.detectChanges();
      const banner = fixture.debugElement.query(By.css('.error-banner'));
      expect(banner).toBeTruthy();
      expect(banner.nativeElement.textContent).toContain(expected);
    });

    it('should hide when no errors', () => {
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeNull();
    });

    it('should show and hide reactively', () => {
      gatewayErrorSig.set('Error appeared');
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeTruthy();

      gatewayErrorSig.set(null);
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeNull();
    });

    it('should dismiss the error when dismiss button is clicked', () => {
      gatewayErrorSig.set('Gateway failed');
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeTruthy();

      fixture.debugElement.query(By.css('.error-banner button')).nativeElement.click();
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeNull();
    });
  });

  describe('template rendering', () => {
    it('should render gateway table stub, and alongside error banner when error is set', () => {
      expect(getTable()).toBeTruthy();

      gatewayErrorSig.set('Something went wrong');
      fixture.detectChanges();
      expect(fixture.debugElement.query(By.css('.error-banner'))).toBeTruthy();
      expect(getTable()).toBeTruthy();
    });
  });

  describe('input bindings', () => {
    it('should pass all service signals and actionMode to gateway table', () => {
      (managerServiceMock.gatewayList as WritableSignal<Gateway[]>).set([mockGateway]);
      (managerServiceMock.sensorList as WritableSignal<Sensor[]>).set([mockSensor]);
      (managerServiceMock.gatewayTotal as WritableSignal<number>).set(25);
      (managerServiceMock.gatewayPageIndex as WritableSignal<number>).set(2);
      (managerServiceMock.gatewayLimit as WritableSignal<number>).set(5);
      (managerServiceMock.sensorTotal as WritableSignal<number>).set(10);
      (managerServiceMock.sensorPageIndex as WritableSignal<number>).set(1);
      (managerServiceMock.sensorLimit as WritableSignal<number>).set(15);
      (managerServiceMock.gatewayLoading as WritableSignal<boolean>).set(true);
      (managerServiceMock.sensorLoading as WritableSignal<boolean>).set(true);
      (managerServiceMock.expandedGateway as WritableSignal<Gateway | null>).set(mockGateway);
      fixture.detectChanges();

      const table = getTable().componentInstance as StubGatewayTable;
      expect(table.actionMode()).toBe('manage');
      expect(table.gateways()).toEqual([mockGateway]);
      expect(table.sensors()).toEqual([mockSensor]);
      expect(table.gatewayTotal()).toBe(25);
      expect(table.gatewayPageIndex()).toBe(2);
      expect(table.gatewayLimit()).toBe(5);
      expect(table.sensorTotal()).toBe(10);
      expect(table.sensorPageIndex()).toBe(1);
      expect(table.sensorLimit()).toBe(15);
      expect(table.gatewayLoading()).toBe(true);
      expect(table.sensorLoading()).toBe(true);
      expect(table.expandedGateway()).toEqual(mockGateway);
    });
  });

  describe('output events', () => {
    it('should call toggleExpandedGateway when table emits expandedGatewayChange', () => {
      getTable().triggerEventHandler('expandedGatewayChange', mockGateway);
      expect(managerServiceMock.toggleExpandedGateway).toHaveBeenCalledWith(mockGateway);
    });

    it('should call changeGatewayPage when table emits gatewayPageChange', () => {
      getTable().triggerEventHandler('gatewayPageChange', {
        pageIndex: 2,
        pageSize: 25,
        length: 100,
      });
      expect(managerServiceMock.changeGatewayPage).toHaveBeenCalledWith(2, 25);
    });

    it('should call changeSensorPage when table emits sensorPageChange', () => {
      getTable().triggerEventHandler('sensorPageChange', {
        pageIndex: 1,
        pageSize: 10,
        length: 50,
      });
      expect(managerServiceMock.changeSensorPage).toHaveBeenCalledWith(1, 10);
    });
  });

  describe('onCreateGateway', () => {
    it('should open CreateGatewayDialog and refresh on result', () => {
      mockDialog(true);
      getTable().triggerEventHandler('gatewayCreateRequested');
      expect(dialogMock.open).toHaveBeenCalledWith(CreateGatewayDialog);
      expect(managerServiceMock.refreshGateways).toHaveBeenCalledOnce();
      expect(snackBarMock.open).toHaveBeenCalledWith('Gateway creato con successo', 'Close', {
        duration: 3000,
      });
    });

    it.each([undefined, null, false, ''])('should not refresh when dialog returns %s', (result) => {
      mockDialog(result);
      getTable().triggerEventHandler('gatewayCreateRequested');
      expect(managerServiceMock.refreshGateways).not.toHaveBeenCalled();
      expect(snackBarMock.open).not.toHaveBeenCalled();
    });
  });

  describe('onDeleteGateway', () => {
    it('should open ConfirmDeleteDialog and delete on confirm', () => {
      mockDialog(true);
      getTable().triggerEventHandler('gatewayDeleteRequested', mockGateway);
      expect(dialogMock.open).toHaveBeenCalledWith(ConfirmDeleteDialog, {
        data: {
          title: 'Elimina Gateway',
          message: `Sei sicuro di voler eliminare il gateway "${mockGateway.name}"?`,
        },
      });
      expect(managerServiceMock.deleteGateway).toHaveBeenCalledWith(mockGateway);
      expect(snackBarMock.open).toHaveBeenCalledWith('Gateway eliminato con successo', 'Close', {
        duration: 3000,
      });
    });

    it.each([undefined, false, null])('should not delete when dialog returns %s', (result) => {
      mockDialog(result);
      getTable().triggerEventHandler('gatewayDeleteRequested', mockGateway);
      expect(managerServiceMock.deleteGateway).not.toHaveBeenCalled();
      expect(snackBarMock.open).not.toHaveBeenCalled();
    });
  });

  describe('onCreateSensor', () => {
    it('should open CreateSensorDialog with gateway data and refresh on result', () => {
      mockDialog(true);
      getTable().triggerEventHandler('sensorCreateRequested', mockGateway);
      expect(dialogMock.open).toHaveBeenCalledWith(CreateSensorDialog, {
        data: { id: 'gw-1', name: 'Gateway 1' },
      });
      expect(managerServiceMock.refreshSensors).toHaveBeenCalledWith('gw-1');
      expect(snackBarMock.open).toHaveBeenCalledWith('Sensore creato con successo', 'Close', {
        duration: 3000,
      });
    });

    it.each([undefined, null, false])('should not refresh when dialog returns %s', (result) => {
      mockDialog(result);
      getTable().triggerEventHandler('sensorCreateRequested', mockGateway);
      expect(managerServiceMock.refreshSensors).not.toHaveBeenCalled();
      expect(snackBarMock.open).not.toHaveBeenCalled();
    });
  });

  describe('onDeleteSensor', () => {
    it('should open ConfirmDeleteDialog and delete on confirm', () => {
      mockDialog(true);
      getTable().triggerEventHandler('sensorDeleteRequested', mockSensor);
      expect(dialogMock.open).toHaveBeenCalledWith(ConfirmDeleteDialog, {
        data: {
          title: 'Elimina Sensore',
          message: `Sei sicuro di voler eliminare il sensore "${mockSensor.name}"?`,
        },
      });
      expect(managerServiceMock.deleteSensor).toHaveBeenCalledWith(mockSensor);
      expect(snackBarMock.open).toHaveBeenCalledWith('Sensore eliminato con successo', 'Close', {
        duration: 3000,
      });
    });

    it.each([undefined, false, null])('should not delete when dialog returns %s', (result) => {
      mockDialog(result);
      getTable().triggerEventHandler('sensorDeleteRequested', mockSensor);
      expect(managerServiceMock.deleteSensor).not.toHaveBeenCalled();
      expect(snackBarMock.open).not.toHaveBeenCalled();
    });
  });
});
