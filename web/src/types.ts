export interface Secret {
  id: string;
  key: string;
  value: string;
  consumerId: string;
}

export interface Consumer {
  id: string;
  name: string;
}

export interface ConsumerSecrets {
  consumerId: string;
  secrets: Secret[];
}

export interface ApiResponse<T> {
  status: number;
  message: string;
  data?: T;
}
