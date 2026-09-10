export interface Secret {
  id: string;
  key: string;
  value: string;
  consumerId: string;
  isSecret: boolean;
}

export interface Consumer {
  id: string;
  name: string;
  apikey: string;
  roleId: string;
  active: boolean;
}

export interface Role {
  id: string;
  name: string;
}

export interface LoginResponse {
  token: string;
  username: string;
  expiresAt: number;
}

export interface ApiResponse<T> {
  status: number;
  message: string;
  data?: T;
}