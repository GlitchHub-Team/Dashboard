import { FieldError } from './field-error.model';

export interface ApiError {
  status: string;
  message: string;
  errors: FieldError[];
}
