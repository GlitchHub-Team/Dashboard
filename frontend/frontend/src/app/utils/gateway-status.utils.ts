import { EnumMapper } from "./enum.utils";
import { GatewayStatus } from "../models/gateway-status.enum";

export const gatewayStatusMapper = new EnumMapper<GatewayStatus, string>(
  {
    [GatewayStatus.ACTIVE]: "active",
    [GatewayStatus.INACTIVE]: "inactive",
    [GatewayStatus.DECOMMISSIONED]: "decommissioned",
  },
  GatewayStatus.INACTIVE,
);