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

// ── Auth ──────────────────────────────────────────────────────────────────────

export interface TokenPair {
	access_token: string;
	expires_in: number;
	refresh_token?: string;
}

export interface UserRole {
	role_id: number;
	role_name: string;
	branch_id?: number;
}

export interface MeResponse {
	id: number;
	organization_id: number;
	name: string;
	email: string;
	roles: UserRole[];
}

// ── Organization ──────────────────────────────────────────────────────────────

export interface Organization {
	id: number;
	name: string;
	slug: string;
	logo_url?: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

// ── Branch ────────────────────────────────────────────────────────────────────

export interface Branch {
	id: number;
	organization_id: number;
	name: string;
	slug: string;
	address?: string;
	phone?: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface CreateBranchPayload {
	name: string;
	slug: string;
	address?: string;
	phone?: string;
}

export interface UpdateBranchPayload {
	name?: string;
	slug?: string;
	address?: string;
	phone?: string;
	is_active?: boolean;
}

// ── Menu ──────────────────────────────────────────────────────────────────────

export interface MenuCategory {
	id: number;
	organization_id: number;
	name: string;
	description?: string;
	sort_order: number;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface MenuItem {
	id: number;
	organization_id: number;
	category_id: number;
	name: string;
	description?: string;
	base_price: string;
	image_url?: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface MenuBranchSetting {
	id: number;
	menu_item_id: number;
	branch_id: number;
	price_override?: string;
	is_available: boolean;
	updated_at: string;
}

// ── HTTP helper ───────────────────────────────────────────────────────────────

async function request<T>(
	path: string,
	init?: RequestInit & { token?: string }
): Promise<ApiResponse<T>> {
	try {
		const headers: Record<string, string> = {
			'Content-Type': 'application/json',
			...(init?.headers as Record<string, string>)
		};
		if (init?.token) {
			const bearerScheme = 'Bearer ';
			headers['Authorization'] = bearerScheme + init.token;
		}
		const res = await fetch(`${API_BASE_URL}${path}`, {
			...init,
			headers
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

// ── API surface ───────────────────────────────────────────────────────────────

export const api = {
	health(): Promise<ApiResponse<HealthStatus>> {
		return request<HealthStatus>('/health');
	},
	info(): Promise<ApiResponse<AppInfo>> {
		return request<AppInfo>('/api/v1/info');
	},

	// Auth
	login(email: string, password: string): Promise<ApiResponse<TokenPair>> {
		return request<TokenPair>('/api/v1/auth/login', {
			method: 'POST',
			body: JSON.stringify({ email, password })
		});
	},
	refresh(refreshToken: string): Promise<ApiResponse<TokenPair>> {
		return request<TokenPair>('/api/v1/auth/refresh', {
			method: 'POST',
			body: JSON.stringify({ refresh_token: refreshToken })
		});
	},
	logout(token: string, refreshToken?: string): Promise<ApiResponse<{ message: string }>> {
		return request('/api/v1/auth/logout', {
			method: 'POST',
			token,
			body: JSON.stringify({ refresh_token: refreshToken })
		});
	},
	me(token: string): Promise<ApiResponse<MeResponse>> {
		return request<MeResponse>('/api/v1/auth/me', { token });
	},

	// Organization
	organization(token: string): Promise<ApiResponse<Organization>> {
		return request<Organization>('/api/v1/organization', { token });
	},

	// Branches
	branches: {
		list(token: string): Promise<ApiResponse<{ data: Branch[] }>> {
			return request<{ data: Branch[] }>('/api/v1/branches', { token });
		},
		get(token: string, id: number): Promise<ApiResponse<Branch>> {
			return request<Branch>(`/api/v1/branches/${id}`, { token });
		},
		create(token: string, payload: CreateBranchPayload): Promise<ApiResponse<Branch>> {
			return request<Branch>('/api/v1/branches', {
				method: 'POST',
				token,
				body: JSON.stringify(payload)
			});
		},
		update(token: string, id: number, payload: UpdateBranchPayload): Promise<ApiResponse<Branch>> {
			return request<Branch>(`/api/v1/branches/${id}`, {
				method: 'PATCH',
				token,
				body: JSON.stringify(payload)
			});
		}
	},

	// Menu
	menu: {
		categories: {
			list(token: string): Promise<ApiResponse<{ data: MenuCategory[] }>> {
				return request<{ data: MenuCategory[] }>('/api/v1/menu/categories', { token });
			},
			create(
				token: string,
				payload: { name: string; description?: string; sort_order?: number }
			): Promise<ApiResponse<MenuCategory>> {
				return request<MenuCategory>('/api/v1/menu/categories', {
					method: 'POST',
					token,
					body: JSON.stringify(payload)
				});
			},
			update(
				token: string,
				id: number,
				payload: { name?: string; description?: string; sort_order?: number; is_active?: boolean }
			): Promise<ApiResponse<MenuCategory>> {
				return request<MenuCategory>(`/api/v1/menu/categories/${id}`, {
					method: 'PATCH',
					token,
					body: JSON.stringify(payload)
				});
			}
		},
		items: {
			list(
				token: string,
				filters?: { category_id?: number; branch_id?: number; active?: boolean }
			): Promise<ApiResponse<{ data: MenuItem[] }>> {
				const params = new URLSearchParams();
				if (filters?.category_id) params.set('category_id', String(filters.category_id));
				if (filters?.branch_id) params.set('branch_id', String(filters.branch_id));
				if (filters?.active) params.set('active', '1');
				const qs = params.toString() ? '?' + params.toString() : '';
				return request<{ data: MenuItem[] }>(`/api/v1/menu/items${qs}`, { token });
			},
			get(token: string, id: number): Promise<ApiResponse<MenuItem>> {
				return request<MenuItem>(`/api/v1/menu/items/${id}`, { token });
			},
			create(
				token: string,
				payload: {
					category_id: number;
					name: string;
					description?: string;
					base_price: string;
					image_url?: string;
				}
			): Promise<ApiResponse<MenuItem>> {
				return request<MenuItem>('/api/v1/menu/items', {
					method: 'POST',
					token,
					body: JSON.stringify(payload)
				});
			},
			update(
				token: string,
				id: number,
				payload: {
					category_id?: number;
					name?: string;
					description?: string;
					base_price?: string;
					image_url?: string;
					is_active?: boolean;
				}
			): Promise<ApiResponse<MenuItem>> {
				return request<MenuItem>(`/api/v1/menu/items/${id}`, {
					method: 'PATCH',
					token,
					body: JSON.stringify(payload)
				});
			},
			branchSettings(token: string, itemId: number): Promise<ApiResponse<{ data: MenuBranchSetting[] }>> {
				return request<{ data: MenuBranchSetting[] }>(`/api/v1/menu/items/${itemId}/branches`, {
					token
				});
			},
			upsertBranchSetting(
				token: string,
				itemId: number,
				branchId: number,
				payload: { price_override?: string; is_available: boolean }
			): Promise<ApiResponse<MenuBranchSetting>> {
				return request<MenuBranchSetting>(
					`/api/v1/menu/items/${itemId}/branches/${branchId}`,
					{
						method: 'PUT',
						token,
						body: JSON.stringify(payload)
					}
				);
			}
		}
	}
};

