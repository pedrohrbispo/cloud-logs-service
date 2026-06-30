import { apiClient } from './client';
import type { ListProvidersData, ProviderInfo } from './types';

export async function listProviders(): Promise<ProviderInfo[]> {
  const data = await apiClient.get<ListProvidersData>('/providers');
  return data.providers;
}
