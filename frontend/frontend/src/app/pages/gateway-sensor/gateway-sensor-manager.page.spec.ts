import { describe, it, expect, vi, beforeEach } from 'vitest';
import { signal, WritableSignal, NO_ERRORS_SCHEMA } from '@angular/core';
import { TestBed, ComponentFixture } from '@angular/core/testing';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { PageEvent } from '@angular/material/paginator';
import { of } from 'rxjs';

import { GatewaySensorManagerPage } from '../../pages/gateway-sensor/gateway-sensor-manager.page';
import { GatewaySensorManagerService } from '../../services/gateway-sensor-manager/gateway-sensor-manager.service';
import { Gateway } from '../../models/gateway/gateway.model';
import { GatewayStatus } from '../../models/gateway/gateway-status.enum';
import { Sensor } from '../../models/sensor/sensor.model';
import { SensorProfiles } from '../../models/sensor/sensor-profiles.enum';

describe('GatewaySensorManagerPage', () => {
  let fixture: ComponentFixture<GatewaySensorManagerPage>;
  let component: GatewaySensorManagerPage;

  let gatewayErrorSig: WritableSignal<string | null>;
  let sensorErrorSig: WritableSignal<string | null>;

  const mockGateway: Gateway = {
    id: 'gw-1',
    tenantId: 'tenant-01',
    name: 'Gateway 1',
    status: GatewayStatus.ONLINE,
  };

  const mockSensor: Sensor = {
    id: 'sensor-1',
    gatewayId: 'gw-1',
    name: 'Heart Rate Sensor',
    profile: SensorProfiles.HEART_RATE_SERVICE,
    dataInterval: 1000,
  };

  let managerServiceMock: any;
  let dialogMock: any;
  let snackBarMock: any;

  beforeEach(async () => {
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
      createGateway: vi.fn(),
      deleteGateway: vi.fn(),
      createSensor: vi.fn(),
      deleteSensor: vi.fn(),
    };

    dialogMock = { open: vi.fn() };
    snackBarMock = { open: vi.fn() };

    await TestBed.configureTestingModule({
      imports: [GatewaySensorManagerPage],
      schemas: [NO_ERRORS_SCHEMA],
      providers: [
        { provide: GatewaySensorManagerService, useValue: managerServiceMock },
        { provide: MatDialog, useValue: dialogMock },
        { provide: MatSnackBar, useValue: snackBarMock },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(GatewaySensorManagerPage);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  const query = (selector: string): HTMLElement | null =>
    fixture.nativeElement.querySelector(selector);

  function mockDialog(result: any): void {
    dialogMock.open.mockReturnValue({ afterClosed: () => of(result) });
  }

  it('should call loadGateways on init', () => {
    expect(managerServiceMock.loadGateways).toHaveBeenCalledOnce();
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

      const banner = query('.error-banner');
      expect(banner).not.toBeNull();
      expect(banner!.textContent).toContain(expected);
    });

    it('should hide and show banner reactively as errors change', () => {
      expect(query('.error-banner')).toBeNull();

      gatewayErrorSig.set('Error appeared');
      fixture.detectChanges();
      expect(query('.error-banner')).not.toBeNull();

      gatewayErrorSig.set(null);
      fixture.detectChanges();
      expect(query('.error-banner')).toBeNull();
    });
  });

  describe('template rendering', () => {
    it('should render the gateway table', () => {
      expect(query('app-dashboard-gateway-table')).not.toBeNull();
    });

    it('should render error banner and gateway table simultaneously', () => {
      gatewayErrorSig.set('Something went wrong');
      fixture.detectChanges();

      expect(query('.error-banner')).not.toBeNull();
      expect(query('app-dashboard-gateway-table')).not.toBeNull();
    });
  });

  it('should call toggleExpandedGateway on the service', () => {
    component['onExpandedGatewayChange'](mockGateway);
    expect(managerServiceMock.toggleExpandedGateway).toHaveBeenCalledWith(mockGateway);
  });

  it('should call changeGatewayPage with pageIndex and pageSize', () => {
    component['onGatewayPageChange']({ pageIndex: 2, pageSize: 25, length: 100 } as PageEvent);
    expect(managerServiceMock.changeGatewayPage).toHaveBeenCalledWith(2, 25);
  });

  it('should call changeSensorPage with pageIndex and pageSize', () => {
    component['onSensorPageChange']({ pageIndex: 1, pageSize: 10, length: 50 } as PageEvent);
    expect(managerServiceMock.changeSensorPage).toHaveBeenCalledWith(1, 10);
  });

  describe('onCreateGateway', () => {
    it('should open CreateGatewayDialog and call createGateway on result', () => {
      const newGateway = { name: 'New Gateway' };
      mockDialog(newGateway);

      component['onCreateGateway']();

      expect(dialogMock.open).toHaveBeenCalledOnce();
      expect(managerServiceMock.createGateway).toHaveBeenCalledWith(newGateway);
      expect(snackBarMock.open).toHaveBeenCalledWith('Gateway created', 'Close', {
        duration: 3000,
      });
    });

    it.each([undefined, null, ''])(
      'should not call createGateway when dialog returns %s',
      (result) => {
        mockDialog(result);
        component['onCreateGateway']();

        expect(managerServiceMock.createGateway).not.toHaveBeenCalled();
        expect(snackBarMock.open).not.toHaveBeenCalled();
      },
    );
  });

  describe('onDeleteGateway', () => {
    it('should open ConfirmDeleteDialog with gateway data and delete on confirm', () => {
      mockDialog(true);
      component['onDeleteGateway'](mockGateway);

      expect(dialogMock.open).toHaveBeenCalledWith(expect.anything(), {
        data: { entityName: 'Gateway 1', entityType: 'gateway' },
      });
      expect(managerServiceMock.deleteGateway).toHaveBeenCalledWith(mockGateway);
      expect(snackBarMock.open).toHaveBeenCalledWith('Gateway deleted', 'Close', {
        duration: 3000,
      });
    });

    it.each([undefined, false])(
      'should not call deleteGateway when dialog returns %s',
      (result) => {
        mockDialog(result);
        component['onDeleteGateway'](mockGateway);

        expect(managerServiceMock.deleteGateway).not.toHaveBeenCalled();
        expect(snackBarMock.open).not.toHaveBeenCalled();
      },
    );
  });

  describe('onCreateSensor', () => {
    it('should open CreateSensorDialog with gatewayId and call createSensor on result', () => {
      const newSensor = { name: 'New Sensor', profile: SensorProfiles.HEART_RATE_SERVICE };
      mockDialog(newSensor);

      component['onCreateSensor'](mockGateway);

      expect(dialogMock.open).toHaveBeenCalledWith(expect.anything(), {
        data: { gatewayId: 'gw-1' },
      });
      expect(managerServiceMock.createSensor).toHaveBeenCalledWith('gw-1', newSensor);
      expect(snackBarMock.open).toHaveBeenCalledWith('Sensor created', 'Close', { duration: 3000 });
    });

    it.each([undefined, null])('should not call createSensor when dialog returns %s', (result) => {
      mockDialog(result);
      component['onCreateSensor'](mockGateway);

      expect(managerServiceMock.createSensor).not.toHaveBeenCalled();
      expect(snackBarMock.open).not.toHaveBeenCalled();
    });
  });

  describe('onDeleteSensor', () => {
    it('should open ConfirmDeleteDialog with sensor data and delete on confirm', () => {
      mockDialog(true);
      component['onDeleteSensor'](mockSensor);

      expect(dialogMock.open).toHaveBeenCalledWith(expect.anything(), {
        data: { entityName: 'Heart Rate Sensor', entityType: 'sensor' },
      });
      expect(managerServiceMock.deleteSensor).toHaveBeenCalledWith(mockSensor);
      expect(snackBarMock.open).toHaveBeenCalledWith('Sensor deleted', 'Close', { duration: 3000 });
    });

    it.each([undefined, false])('should not call deleteSensor when dialog returns %s', (result) => {
      mockDialog(result);
      component['onDeleteSensor'](mockSensor);

      expect(managerServiceMock.deleteSensor).not.toHaveBeenCalled();
      expect(snackBarMock.open).not.toHaveBeenCalled();
    });
  });
});
