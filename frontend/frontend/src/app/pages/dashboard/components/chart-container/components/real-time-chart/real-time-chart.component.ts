import { Component, computed, input } from '@angular/core';
import { Chart, registerables, ChartData, ChartOptions } from 'chart.js';
import { BaseChartDirective } from 'ng2-charts';

import { SensorReading } from '../../../../../../models/sensor-data/sensor-reading.model';
import { Sensor } from '../../../../../../models/sensor/sensor.model';
import { getSensorProfileDisplay } from '../../../../../../models/chart/sensor-profile-display.model';

Chart.register(...registerables);

@Component({
  selector: 'app-real-time-chart',
  imports: [BaseChartDirective],
  templateUrl: './real-time-chart.component.html',
  styleUrl: './real-time-chart.component.css',
})
export class RealTimeChartComponent {
  public readings = input.required<SensorReading[]>();
  public sensor = input.required<Sensor>();

  protected readonly profileDisplay = computed(() =>
    getSensorProfileDisplay(this.sensor().profile),
  );

  protected readonly chartData = computed<ChartData<'line'>>(() => {
    const readings = this.readings();
    return {
      labels: readings.map((r) => new Date(r.timestamp).toLocaleTimeString()),
      datasets: [
        {
          label: this.profileDisplay().label,
          data: readings.map((r) => r.value),
          borderColor: '#4caf50',
          backgroundColor: 'rgba(76, 175, 80, 0.1)',
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
    animation: false,
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
}
