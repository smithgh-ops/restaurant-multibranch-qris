const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

export interface ApiResponse<T = unknown> {
	data?: T;
	error?: string;
	message?: string;
}

export interface AppInfo {
	name: string;
	version: string;
	env: string;
}

export interface HealthStatus {
	status: string;
	timestamp: string;
}

async function request<T>(path: string, init?: RequestInit): Promise<ApiResponse<T>> {
	try {
		const res = await fetch(`${API_BASE_URL}${path}`, {
			headers: {
				'Content-Type': 'application/json',
				...init?.headers
			},
			...init
		});
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return { error: body.error ?? `HTTP ${res.status}` };
		}
		const data: T = await res.json();
		return { data };
	} catch (err) {
		return { error: err instanceof Error ? err.message : 'Network error' };
	}
}

export const api = {
	health(): Promise<ApiResponse<HealthStatus>> {
		return request<HealthStatus>('/health');
	},
	info(): Promise<ApiResponse<AppInfo>> {
		return request<AppInfo>('/api/v1/info');
	}
};
