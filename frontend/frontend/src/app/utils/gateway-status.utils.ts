import { EnumMapper } from "./enum.utils";
import { GatewayStatus } from "../models/gateway-status.enum";

export const gatewayStatusMapper = new EnumMapper<GatewayStatus, string>(
  {
    [GatewayStatus.ACTIVE]: "active",
    [GatewayStatus.INACTIVE]: "inactive",
    [GatewayStatus.COMMISSIONED]: "commissioned",
    [GatewayStatus.DECOMMISSIONED]: "decommissioned",
    [GatewayStatus.INTERRUPTED]: "interrupted",
  },
  GatewayStatus.INACTIVE,
);