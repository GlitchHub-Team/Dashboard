import { FieldError } from './field-error.model';

export interface ApiError {
  status: number;
  message: string;
  errors?: FieldError[];
}
