export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

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

// ── Table ─────────────────────────────────────────────────────────────────────

export interface DiningArea {
	id: number;
	branch_id: number;
	name: string;
	description?: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface RestaurantTable {
	id: number;
	dining_area_id: number;
	branch_id: number;
	table_number: string;
	capacity: number;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface CreateTablePayload {
	dining_area_id: number;
	table_number: string;
	capacity?: number;
}

export interface UpdateTablePayload {
	table_number?: string;
	capacity?: number;
	is_active?: boolean;
}

// ── Order ─────────────────────────────────────────────────────────────────────

export type OrderStatus =
	'pending' | 'confirmed' | 'preparing' | 'ready' | 'completed' | 'cancelled';
export type OrderType = 'dine_in' | 'takeaway' | 'delivery';

export interface OrderItem {
	id: number;
	order_id: number;
	menu_item_id: number;
	item_name: string;
	unit_price: string;
	quantity: number;
	subtotal: string;
	notes?: string;
}

export interface Order {
	id: number;
	branch_id: number;
	table_id?: number;
	order_code: string;
	order_type: OrderType;
	status: OrderStatus;
	notes?: string;
	subtotal: string;
	tax_amount: string;
	service_charge: string;
	total_amount: string;
	created_by?: number;
	created_at: string;
	updated_at: string;
	items: OrderItem[];
}

export interface CreateOrderPayload {
	branch_id: number;
	table_id?: number;
	order_type?: OrderType;
	notes?: string;
	items: { menu_item_id: number; quantity: number; notes?: string }[];
}

// ── Payment ───────────────────────────────────────────────────────────────────

export type PaymentStatus = 'pending' | 'paid' | 'failed' | 'expired';

export interface Payment {
	id: number;
	order_id: number;
	gateway_config_id?: number;
	gateway_invoice_id: string;
	payment_method: string;
	amount: string;
	status: PaymentStatus;
	paid_at?: string;
	qr_code_url?: string;
	qr_string?: string;
	expiry_at?: string;
	created_at: string;
	updated_at: string;
}

// ── KDS ───────────────────────────────────────────────────────────────────────

export type KitchenTicketStatus = 'queued' | 'in_progress' | 'done' | 'cancelled';

export interface KitchenStation {
	id: number;
	branch_id: number;
	name: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface KitchenTicket {
	id: number;
	order_id: number;
	station_id?: number;
	station_name?: string;
	order_item_id: number;
	branch_id: number;
	table_id?: number;
	order_code: string;
	item_name: string;
	quantity: number;
	notes?: string;
	status: KitchenTicketStatus;
	priority: number;
	started_at?: string;
	completed_at?: string;
	created_at: string;
	updated_at: string;
}

// ── Reports ───────────────────────────────────────────────────────────────────

export interface SalesSummary {
	total_revenue: string;
	order_count: number;
	avg_order_value: string;
}

export interface DailySales {
	date: string;
	revenue: string;
	order_count: number;
}

export interface SalesReport {
	summary: SalesSummary;
	daily: DailySales[];
}

export interface TopItem {
	menu_item_id: number;
	item_name: string;
	total_qty: number;
	total_revenue: string;
}

// ── Self-order (public QR table) ──────────────────────────────────────────────

export interface QRTableToken {
	id: number;
	table_id: number;
	branch_id: number;
	token: string;
	is_active: boolean;
	expires_at?: string;
	created_at: string;
	updated_at: string;
}

export interface PublicBranch {
	id: number;
	name: string;
	address?: string;
}

export interface PublicTable {
	id: number;
	table_number: string;
	capacity: number;
}

export interface PublicMenuItem {
	id: number;
	category_id: number;
	name: string;
	description?: string;
	price: string;
	image_url?: string;
}

export interface PublicMenuCategory {
	id: number;
	name: string;
	items: PublicMenuItem[];
}

export interface PublicMenuResponse {
	branch: PublicBranch;
	table: PublicTable;
	categories: PublicMenuCategory[];
}

export interface SelfOrderResponse {
	order_id: number;
	order_code: string;
	status: string;
	total_amount: string;
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
			branchSettings(
				token: string,
				itemId: number
			): Promise<ApiResponse<{ data: MenuBranchSetting[] }>> {
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
				return request<MenuBranchSetting>(`/api/v1/menu/items/${itemId}/branches/${branchId}`, {
					method: 'PUT',
					token,
					body: JSON.stringify(payload)
				});
			}
		}
	},

	// Tables
	tables: {
		diningAreas(token: string, branchId: number): Promise<ApiResponse<{ data: DiningArea[] }>> {
			return request<{ data: DiningArea[] }>(`/api/v1/branches/${branchId}/dining-areas`, {
				token
			});
		},
		createDiningArea(
			token: string,
			branchId: number,
			payload: { name: string; description?: string }
		): Promise<ApiResponse<DiningArea>> {
			return request<DiningArea>(`/api/v1/branches/${branchId}/dining-areas`, {
				method: 'POST',
				token,
				body: JSON.stringify(payload)
			});
		},
		list(token: string, branchId: number): Promise<ApiResponse<{ data: RestaurantTable[] }>> {
			return request<{ data: RestaurantTable[] }>(`/api/v1/branches/${branchId}/tables`, { token });
		},
		create(
			token: string,
			branchId: number,
			payload: CreateTablePayload
		): Promise<ApiResponse<RestaurantTable>> {
			return request<RestaurantTable>(`/api/v1/branches/${branchId}/tables`, {
				method: 'POST',
				token,
				body: JSON.stringify(payload)
			});
		},
		update(
			token: string,
			branchId: number,
			tableId: number,
			payload: UpdateTablePayload
		): Promise<ApiResponse<RestaurantTable>> {
			return request<RestaurantTable>(`/api/v1/branches/${branchId}/tables/${tableId}`, {
				method: 'PATCH',
				token,
				body: JSON.stringify(payload)
			});
		}
	},

	// Orders
	orders: {
		list(
			token: string,
			filters?: { branch_id?: number; status?: OrderStatus }
		): Promise<ApiResponse<{ data: Order[] }>> {
			const params = new URLSearchParams();
			if (filters?.branch_id) params.set('branch_id', String(filters.branch_id));
			if (filters?.status) params.set('status', filters.status);
			const qs = params.toString() ? '?' + params.toString() : '';
			return request<{ data: Order[] }>(`/api/v1/orders${qs}`, { token });
		},
		get(token: string, id: number): Promise<ApiResponse<Order>> {
			return request<Order>(`/api/v1/orders/${id}`, { token });
		},
		create(token: string, payload: CreateOrderPayload): Promise<ApiResponse<Order>> {
			return request<Order>('/api/v1/orders', {
				method: 'POST',
				token,
				body: JSON.stringify(payload)
			});
		},
		updateStatus(token: string, id: number, status: OrderStatus): Promise<ApiResponse<Order>> {
			return request<Order>(`/api/v1/orders/${id}/status`, {
				method: 'PATCH',
				token,
				body: JSON.stringify({ status })
			});
		}
	},

	// Payments
	payments: {
		createInvoice(
			token: string,
			orderId: number,
			payload?: { expiry_minutes?: number }
		): Promise<ApiResponse<Payment>> {
			return request<Payment>(`/api/v1/orders/${orderId}/payments/qris-invoice`, {
				method: 'POST',
				token,
				body: JSON.stringify(payload ?? {})
			});
		}
	},

	// Self-order QR table tokens
	selforder: {
		generateToken(token: string, branchId: number, tableId: number): Promise<ApiResponse<QRTableToken>> {
			return request<QRTableToken>(`/api/v1/branches/${branchId}/tables/${tableId}/qr-token`, {
				method: 'POST',
				token
			});
		},
		getToken(token: string, branchId: number, tableId: number): Promise<ApiResponse<QRTableToken>> {
			return request<QRTableToken>(`/api/v1/branches/${branchId}/tables/${tableId}/qr-token`, {
				token
			});
		},
		getMenu(tableToken: string): Promise<ApiResponse<PublicMenuResponse>> {
			return request<PublicMenuResponse>(`/api/v1/public/table/${tableToken}`);
		},
		createOrder(
			tableToken: string,
			payload: {
				notes?: string;
				items: { menu_item_id: number; quantity: number; notes?: string }[];
			}
		): Promise<ApiResponse<SelfOrderResponse>> {
			return request<SelfOrderResponse>(`/api/v1/public/table/${tableToken}/orders`, {
				method: 'POST',
				body: JSON.stringify(payload)
			});
		}
	},

	// KDS
	kds: {
		stations: {
			list(token: string, branchId: number): Promise<ApiResponse<{ data: KitchenStation[] }>> {
				return request<{ data: KitchenStation[] }>(`/api/v1/branches/${branchId}/kds/stations`, {
					token
				});
			},
			create(
				token: string,
				branchId: number,
				payload: { name: string }
			): Promise<ApiResponse<KitchenStation>> {
				return request<KitchenStation>(`/api/v1/branches/${branchId}/kds/stations`, {
					method: 'POST',
					token,
					body: JSON.stringify(payload)
				});
			}
		},
		tickets: {
			list(
				token: string,
				branchId: number,
				filters?: { station_id?: number; status?: KitchenTicketStatus }
			): Promise<ApiResponse<{ data: KitchenTicket[] }>> {
				const params = new URLSearchParams();
				if (filters?.station_id) params.set('station_id', String(filters.station_id));
				if (filters?.status) params.set('status', filters.status);
				const qs = params.toString() ? `?${params.toString()}` : '';
				return request<{ data: KitchenTicket[] }>(`/api/v1/branches/${branchId}/kds/tickets${qs}`, {
					token
				});
			},
			updateStatus(
				token: string,
				branchId: number,
				ticketId: number,
				status: KitchenTicketStatus
			): Promise<ApiResponse<KitchenTicket>> {
				return request<KitchenTicket>(
					`/api/v1/branches/${branchId}/kds/tickets/${ticketId}/status`,
					{
						method: 'PATCH',
						token,
						body: JSON.stringify({ status })
					}
				);
			}
		}
	},

	// Reports
	reports: {
		sales(
			token: string,
			filters?: { branch_id?: number; date_from?: string; date_to?: string }
		): Promise<ApiResponse<SalesReport>> {
			const params = new URLSearchParams();
			if (filters?.branch_id) params.set('branch_id', String(filters.branch_id));
			if (filters?.date_from) params.set('date_from', filters.date_from);
			if (filters?.date_to) params.set('date_to', filters.date_to);
			const qs = params.toString() ? '?' + params.toString() : '';
			return request<SalesReport>(`/api/v1/reports/sales${qs}`, { token });
		},
		topItems(
			token: string,
			filters?: { branch_id?: number; date_from?: string; date_to?: string; limit?: number }
		): Promise<ApiResponse<{ data: TopItem[] }>> {
			const params = new URLSearchParams();
			if (filters?.branch_id) params.set('branch_id', String(filters.branch_id));
			if (filters?.date_from) params.set('date_from', filters.date_from);
			if (filters?.date_to) params.set('date_to', filters.date_to);
			if (filters?.limit) params.set('limit', String(filters.limit));
			const qs = params.toString() ? '?' + params.toString() : '';
			return request<{ data: TopItem[] }>(`/api/v1/reports/top-items${qs}`, { token });
		},
		exportURL(filters?: {
			branch_id?: number;
			date_from?: string;
			date_to?: string;
		}): string {
			const params = new URLSearchParams();
			if (filters?.branch_id) params.set('branch_id', String(filters.branch_id));
			if (filters?.date_from) params.set('date_from', filters.date_from);
			if (filters?.date_to) params.set('date_to', filters.date_to);
			const qs = params.toString() ? '?' + params.toString() : '';
			return `${API_BASE_URL}/api/v1/reports/sales/export${qs}`;
		}
	}
};
