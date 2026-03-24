import { Gateway } from './gateway.model';

export interface CommandRequest {
  gateway: Gateway;
  command: string;
}
