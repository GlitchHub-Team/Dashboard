import { Component, computed, inject, OnDestroy, OnInit, signal } from '@angular/core';
import { MatIcon } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatSnackBar } from '@angular/material/snack-bar';
import { PageEvent } from '@angular/material/paginator';
import { ActivatedRoute, Router } from '@angular/router';

import { DashboardService } from '../../services/dashboard/dashboard.service';
import { UserSessionService } from '../../services/user-session/user-session.service';
import { UserRole } from '../../models/user/user-role.enum';
import { DashboardGatewayTableComponent } from './components/dashboard-gateway-table/dashboard-gateway-table.component';
import { DashboardSensorTableComponent } from './components/dashboard-sensor-table/dashboard-sensor-table.component';
import { ChartContainerComponent } from './components/chart-container/chart-container.component';
import { Gateway } from '../../models/gateway/gateway.model';
import { ChartRequest } from '../../models/chart/chart-request.model';

@Component({
  selector: 'app-dashboard',
  imports: [
    DashboardGatewayTableComponent,
    DashboardSensorTableComponent,
    ChartContainerComponent,
    MatIcon,
    MatButtonModule,
  ],
  templateUrl: './dashboard.page.html',
  styleUrl: './dashboard.page.css',
})
export class DashboardPage implements OnInit, OnDestroy {
  private readonly dashboardService = inject(DashboardService);
  private readonly snackBar = inject(MatSnackBar);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly userSession = inject(UserSessionService);

  protected readonly gatewayList = this.dashboardService.gatewayList;
  protected readonly gatewayTotal = this.dashboardService.gatewayTotal;
  protected readonly gatewayPageIndex = this.dashboardService.gatewayPageIndex;
  protected readonly gatewayLimit = this.dashboardService.gatewayLimit;
  protected readonly gatewayLoading = this.dashboardService.gatewayLoading;

  protected readonly sensorList = this.dashboardService.sensorList;
  protected readonly sensorTotal = this.dashboardService.sensorTotal;
  protected readonly sensorPageIndex = this.dashboardService.sensorPageIndex;
  protected readonly sensorLimit = this.dashboardService.sensorLimit;
  protected readonly sensorLoading = this.dashboardService.sensorLoading;

  protected readonly expandedGateway = this.dashboardService.expandedGateway;
  protected readonly selectedChart = this.dashboardService.selectedChart;
  protected readonly canSendCommands = this.dashboardService.canSendCommands;
  protected readonly error = computed(
    () => this.dashboardService.gatewayError() ?? this.dashboardService.sensorError(),
  );

  // Riceve sicuramente il role da app-shell
  protected readonly currentRole = this.userSession.currentUser()!.role;
  protected readonly UserRole = UserRole;

  // Utilizzato solo quando si fa impersonation di un tenant da SUPER_ADMIN, altrimenti è null e viene ignorato
  protected readonly activeTenantId = signal<string | null>(null);

  // In base al tenant impersonato mostra la rispettiva dashboard
  // Altrimenti se non si è SUPER_ADMIN mostra la dashboard del tenant dell'utente
  public ngOnInit(): void {
    if (this.currentRole === UserRole.SUPER_ADMIN) {
      this.route.queryParams.subscribe((params) => {
        const tenantId: string | undefined = params['tenantId'];
        this.activeTenantId.set(tenantId ?? null);
        this.dashboardService.loadDashboard(tenantId);
      });
    } else {
      const tenantId = this.userSession.currentUser()?.tenantId;
      this.dashboardService.loadDashboard(tenantId);
    }
  }

  protected onBackToTenants(): void {
    this.router.navigate(['/tenant-management']);
  }

  public ngOnDestroy(): void {
    this.dashboardService.closeChart();
  }

  protected onExpandedGatewayChange(gateway: Gateway): void {
    this.dashboardService.toggleExpandedGateway(gateway);
  }

  protected onGatewayPageChange(event: PageEvent): void {
    this.dashboardService.changeGatewayPage(event.pageIndex, event.pageSize);
  }

  protected onSensorPageChange(event: PageEvent): void {
    this.dashboardService.changeSensorPage(event.pageIndex, event.pageSize);
  }

  protected onCommandRequested(result: boolean): void {
    if (result) {
      this.snackBar.open('Command sent successfully', 'Close', {
        duration: 3000,
      });
    }
  }

  protected onChartOpen(request: ChartRequest): void {
    this.dashboardService.openChart(request);
  }

  protected onChartClosed(): void {
    this.dashboardService.closeChart();
  }
}
