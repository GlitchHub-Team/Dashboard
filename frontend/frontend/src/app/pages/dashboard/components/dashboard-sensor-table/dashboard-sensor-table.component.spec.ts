import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DashboardSensorTableComponent } from './dashboard-sensor-table.component';
import { Sensor } from '../../../../models/sensor.model';
import { SensorProfiles } from '../../../../models/sensor-profiles.enum';
import { ChartType } from '../../../../models/chart-type.enum';
import { ChartRequest } from '../../../../models/chart-request.model';

describe('DashboardSensorTableComponent', () => {
  let component: DashboardSensorTableComponent;
  let fixture: ComponentFixture<DashboardSensorTableComponent>;

  const mockSensors: Sensor[] = [
    {
      id: '1',
      gatewayId: 'gw-1',
      name: 'Temperature',
      profile: SensorProfiles.HEALTH_THERMOMETER_SERVICE,
    },
    {
      id: '2',
      gatewayId: 'gw-2',
      name: 'Humidity',
      profile: SensorProfiles.ENVIRONMENTAL_SENSING_SERVICE,
    },
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DashboardSensorTableComponent],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardSensorTableComponent);
    component = fixture.componentInstance;

    // Required input
    fixture.componentRef.setInput('sensors', mockSensors);
    fixture.detectChanges();
  });

  describe('initial state', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should have correct displayed columns', () => {
      expect(component['displayedColumns']).toEqual([
        'id',
        'gatewayId',
        'name',
        'profile',
        'actions',
      ]);
    });

    it('should expose ChartType enum', () => {
      expect(component['ChartType']).toBe(ChartType);
    });
  });

  describe('inputs', () => {
    it('should accept sensors', () => {
      expect(component.sensors()).toEqual(mockSensors);
    });

    it('should accept empty sensors array', () => {
      fixture.componentRef.setInput('sensors', []);
      fixture.detectChanges();

      expect(component.sensors()).toEqual([]);
    });

    it('should default loading to undefined', () => {
      expect(component.loading()).toBeUndefined();
    });

    it('should accept loading input', () => {
      fixture.componentRef.setInput('loading', true);
      fixture.detectChanges();

      expect(component.loading()).toBe(true);
    });
  });

  describe('onViewChart', () => {
    it('should emit chartRequested with correct payload', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      const sensor = mockSensors[0];
      component['onViewChart'](sensor, ChartType.HISTORIC);

      const expected: ChartRequest = {
        sensor,
        chartType: ChartType.HISTORIC,
        timeInterval: null!,
      };
      expect(spy).toHaveBeenCalledWith(expected);
    });

    it('should emit with different chart types', () => {
      const spy = vi.fn();
      component.chartRequested.subscribe(spy);

      const sensor = mockSensors[1];

      component['onViewChart'](sensor, ChartType.REALTIME);

      expect(spy).toHaveBeenCalledWith(
        expect.objectContaining({
          sensor,
          chartType: ChartType.REALTIME,
        }),
      );
    });
  });
});
