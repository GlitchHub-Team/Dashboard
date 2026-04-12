import { Component, computed, effect, input, signal } from '@angular/core';
import { Chart, registerables, ChartData, ChartOptions } from 'chart.js';
import { BaseChartDirective } from 'ng2-charts';
import { MatIcon } from '@angular/material/icon';
import { MatIconButton } from '@angular/material/button';
import { MatSlider, MatSliderThumb } from '@angular/material/slider';
import { MatFormField, MatLabel } from '@angular/material/form-field';
import { MatSelect, MatOption } from '@angular/material/select';

import { SensorReading } from '../../../../../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../../../../../models/sensor-data/field-descriptor.model';
import { Sensor } from '../../../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../../../models/sensor/sensor-profiles.enum';
import { SENSOR_VISIBLE_POINTS } from '../../../../../../models/chart/sensor-visible-points.model';

Chart.register(...registerables);

@Component({
  selector: 'app-historic-chart',
  imports: [
    BaseChartDirective,
    MatIcon,
    MatIconButton,
    MatSlider,
    MatSliderThumb,
    MatFormField,
    MatLabel,
    MatSelect,
    MatOption,
  ],
  templateUrl: './historic-chart.component.html',
  styleUrl: './historic-chart.component.css',
})
export class HistoricChartComponent {
  public readings = input.required<SensorReading[]>();
  public sensor = input.required<Sensor>();
  public fields = input.required<FieldDescriptor[]>();
  public samplesPerPacket = input<number | undefined>(undefined);

  protected readonly selectedField = signal<string>('');

  private readonly initFieldEffect = effect(() => {
    const available = this.fields();
    if (available.length > 0 && !this.selectedFieldDescriptor()) {
      this.selectedField.set(available[0].key);
    }
  });

  protected readonly selectedFieldDescriptor = computed(() =>
    this.fields().find((f) => f.key === this.selectedField()),
  );

  protected readonly hasMultipleFields = computed(() => this.fields().length > 1);

  private readonly isEcg = computed(
    () => this.sensor().profile === SensorProfiles.CUSTOM_ECG_SERVICE,
  );

  protected readonly visiblePoints = computed(() => {
    const profile = this.sensor().profile;
    if (profile === SensorProfiles.CUSTOM_ECG_SERVICE) {
      return this.samplesPerPacket() ?? SENSOR_VISIBLE_POINTS[profile] ?? 50;
    }
    return SENSOR_VISIBLE_POINTS[profile] ?? 50;
  });
  protected readonly offset = signal(0);

  protected readonly maxOffset = computed(() =>
    Math.max(0, this.readings().length - this.visiblePoints()),
  );

  protected readonly scrollStep = computed(() => Math.max(1, Math.floor(this.visiblePoints() / 4)));

  protected readonly canScroll = computed(() => this.readings().length > this.visiblePoints());

  protected readonly visibleReadings = computed(() => {
    const all = this.readings();
    const start = this.offset();
    return all.slice(start, start + this.visiblePoints());
  });

  protected readonly chartData = computed<ChartData<'line'>>(() => {
    const readings = this.visibleReadings();
    const field = this.selectedFieldDescriptor();
    if (!field) return { labels: [], datasets: [] };

    return {
      labels: readings.map((r) => new Date(r.timestamp).toLocaleTimeString()),
      datasets: [
        {
          label: field.label,
          data: readings.map((r) => r.value[field.key]),
          borderColor: this.isEcg() ? '#00ff88' : '#3f51b5',
          backgroundColor: this.isEcg() ? 'rgba(0, 255, 136, 0.05)' : 'rgba(63, 81, 181, 0.1)',
          fill: !this.isEcg(),
          tension: this.isEcg() ? 0.2 : 0.3,
          pointRadius: this.isEcg() ? 0 : 2,
          borderWidth: this.isEcg() ? 1.5 : 2,
        },
      ],
    };
  });

  protected readonly chartOptions = computed<ChartOptions<'line'>>(() => {
    const field = this.selectedFieldDescriptor();
    const ecg = this.isEcg();

    return {
      responsive: true,
      maintainAspectRatio: false,
      animation: ecg ? false : { duration: 300 },
      interaction: {
        mode: 'index',
        intersect: false,
      },
      scales: {
        x: {
          display: true,
          title: {
            display: true,
            text: 'Time',
          },
        },
        y: {
          title: {
            display: true,
            text: field ? `${field.label} (${field.unit})` : 'Value',
          },
        },
      },
      plugins: {
        legend: {
          display: true,
          position: 'top',
        },
      },
    };
  });

  protected onFieldChange(key: string): void {
    this.selectedField.set(key);
  }

  protected onOffsetChange(value: number): void {
    this.offset.set(value);
  }

  protected onScrollLeft(): void {
    const step = this.scrollStep();
    this.offset.update((current) => Math.max(0, current - step));
  }

  protected onScrollRight(): void {
    const step = this.scrollStep();
    this.offset.update((current) => Math.min(this.maxOffset(), current + step));
  }
}
