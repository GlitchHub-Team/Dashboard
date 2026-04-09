import { describe, it, expect } from 'vitest';
import { GatewayApiAdapter } from './gateway-api.adapter';
import { GatewayBackend } from '../../models/gateway/gateway-backend.model';
import { Status } from '../../models/gateway-sensor-status.enum';

describe('GatewayApiAdapter', () => {
  const adapter = new GatewayApiAdapter();

  const dto: GatewayBackend = {
    gateway_id: 'gw-1',
    name: 'Gateway 1',
    status: 'active',
    interval: 60,
    tenant_id: 'tenant-01',
    public_identifier: 'pub-id-1',
  };

  describe('fromDTO', () => {
    it.each([
      { field: 'id', dto, expected: 'gw-1' },
      { field: 'name', dto, expected: 'Gateway 1' },
      { field: 'status', dto, expected: Status.ACTIVE },
      { field: 'interval', dto, expected: 60 },
      { field: 'tenantId', dto, expected: 'tenant-01' },
      { field: 'tenantId', dto: { ...dto, tenant_id: undefined }, expected: undefined },
      { field: 'publicIdentifier', dto, expected: 'pub-id-1' },
      { field: 'publicIdentifier', dto: { ...dto, public_identifier: undefined }, expected: undefined },
    ] as const)('should map $field correctly', ({ dto, field, expected }) => {
      expect(adapter.fromDTO(dto)[field]).toEqual(expected);
    });
  });

  describe('fromPaginatedDTO', () => {
    it('should map count, total and all gateways', () => {
      const response = {
        count: 2,
        total: 10,
        gateways: [dto, { ...dto, gateway_id: 'gw-2', status: 'inattivo', tenant_id: undefined }],
      };

      const result = adapter.fromPaginatedDTO(response);

      expect(result.count).toBe(2);
      expect(result.total).toBe(10);
      expect(result.gateways).toHaveLength(2);
      expect(result.gateways[0].id).toBe('gw-1');
      expect(result.gateways[1].id).toBe('gw-2');
      expect(result.gateways[1].status).toBe(Status.INACTIVE);
      expect(result.gateways[1].tenantId).toBeUndefined();
    });

    it('should handle empty array', () => {
      const result = adapter.fromPaginatedDTO({ count: 0, total: 0, gateways: [] });
      expect(result.gateways).toEqual([]);
    });
  });
});
