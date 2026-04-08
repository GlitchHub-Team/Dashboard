import { Component, computed, input, signal } from '@angular/core';
import { Chart, registerables, ChartData, ChartOptions } from 'chart.js';
import { BaseChartDirective } from 'ng2-charts';
import { MatIcon } from '@angular/material/icon';
import { MatIconButton } from '@angular/material/button';
import { MatSlider, MatSliderThumb } from '@angular/material/slider';

import { SensorReading } from '../../../../../../models/sensor-data/sensor-reading.model';
import { Sensor } from '../../../../../../models/sensor/sensor.model';
import { getSensorProfileDisplay } from '../../../../../../models/chart/sensor-profile-display.model';

Chart.register(...registerables);

@Component({
  selector: 'app-historic-chart',
  imports: [BaseChartDirective, MatIcon, MatIconButton, MatSlider, MatSliderThumb],
  templateUrl: './historic-chart.component.html',
  styleUrl: './historic-chart.component.css',
})
export class HistoricChartComponent {
  public readings = input.required<SensorReading[]>();
  public sensor = input.required<Sensor>();

  protected readonly VISIBLE_POINTS = 50;
  protected readonly offset = signal(0);

  protected readonly maxOffset = computed(() =>
    Math.max(0, this.readings().length - this.VISIBLE_POINTS),
  );
  protected readonly scrollStep = computed(() => Math.max(1, Math.floor(this.VISIBLE_POINTS / 4)));
  protected readonly canScroll = computed(() => this.readings().length > this.VISIBLE_POINTS);
  protected readonly visibleReadings = computed(() => {
    const all = this.readings();
    const start = this.offset();
    return all.slice(start, start + this.VISIBLE_POINTS);
  });
  protected readonly profileDisplay = computed(() =>
    getSensorProfileDisplay(this.sensor().profile),
  );

  protected readonly chartData = computed<ChartData<'line'>>(() => {
    const readings = this.visibleReadings();
    return {
      labels: readings.map((r) => new Date(r.timestamp).toLocaleTimeString()),
      datasets: [
        {
          label: this.profileDisplay().label,
          data: readings.map((r) => r.value),
          borderColor: '#3f51b5',
          backgroundColor: 'rgba(63, 81, 181, 0.1)',
          fill: true,
          tension: 0.3,
          pointRadius: 2,
        },
      ],
    };
  });

  protected readonly chartOptions = computed<ChartOptions<'line'>>(() => ({
    responsive: true,
    maintainAspectRatio: false,
    scales: {
      x: {
        title: {
          display: true,
          text: 'Time',
        },
      },
      y: {
        title: {
          display: true,
          text: this.profileDisplay().unit ? 'Value (' + this.profileDisplay().unit + ')' : 'Value',
        },
      },
    },
    plugins: {
      legend: {
        display: true,
        position: 'top',
      },
    },
  }));

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
