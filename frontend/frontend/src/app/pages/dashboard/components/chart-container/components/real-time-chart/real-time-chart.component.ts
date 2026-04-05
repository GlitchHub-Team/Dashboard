import { Component, computed, effect, input, signal } from '@angular/core';
import { Chart, registerables, ChartData, ChartOptions } from 'chart.js';
import { BaseChartDirective } from 'ng2-charts';
import { MatFormField, MatLabel } from '@angular/material/form-field';
import { MatSelect, MatOption } from '@angular/material/select';

import { SensorReading } from '../../../../../../models/sensor-data/sensor-reading.model';
import { FieldDescriptor } from '../../../../../../models/sensor-data/field-descriptor.model';
import { Sensor } from '../../../../../../models/sensor/sensor.model';
import { SensorProfiles } from '../../../../../../models/sensor/sensor-profiles.enum';

Chart.register(...registerables);

@Component({
  selector: 'app-real-time-chart',
  imports: [
    BaseChartDirective,
    MatFormField,
    MatLabel,
    MatSelect,
    MatOption,
  ],
  templateUrl: './real-time-chart.component.html',
  styleUrl: './real-time-chart.component.css',
})
export class RealTimeChartComponent {
  public readings = input.required<SensorReading[]>();
  public sensor = input.required<Sensor>();
  public fields = input.required<FieldDescriptor[]>();

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

  protected readonly chartData = computed<ChartData<'line'>>(() => {
    const readings = this.readings();
    const field = this.selectedFieldDescriptor();
    if (!field) return { labels: [], datasets: [] };

    return {
      labels: readings.map((r) => new Date(r.timestamp).toLocaleTimeString()),
      datasets: [
        {
          label: field.label,
          data: readings.map((r) => r.value[field.key]),
          borderColor: this.isEcg() ? '#00ff88' : '#4caf50',
          backgroundColor: this.isEcg()
            ? 'rgba(0, 255, 136, 0.05)'
            : 'rgba(76, 175, 80, 0.1)',
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
      animation: false,
      scales: {
        x: {
          display: !ecg,
          title: {
            display: !ecg,
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
        tooltip: {
          enabled: !ecg,
        },
      },
    };
  });

  protected onFieldChange(key: string): void {
    this.selectedField.set(key);
  }
}