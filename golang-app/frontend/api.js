// ========================================
// API Client — AuctionHub Frontend
// ========================================

const API_BASE = 'http://localhost:8080/api/v1';

const api = {
    _token: localStorage.getItem('token'),
    _user: JSON.parse(localStorage.getItem('user') || 'null'),

    get token() { return this._token; },
    set token(t) { this._token = t; t ? localStorage.setItem('token', t) : localStorage.removeItem('token'); },

    get user() { return this._user; },
    set user(u) { this._user = u; u ? localStorage.setItem('user', JSON.stringify(u)) : localStorage.removeItem('user'); },

    get isLoggedIn() { return !!this._token; },
    get isAdmin() { return this._user?.role === 'admin'; },

    async request(method, path, body = null) {
        const headers = { 'Content-Type': 'application/json' };
        if (this._token) headers['Authorization'] = `Bearer ${this._token}`;
        const opts = { method, headers };
        if (body) opts.body = JSON.stringify(body);
        const res = await fetch(`${API_BASE}${path}`, opts);
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || data.message || 'Lỗi không xác định');
        return data;
    },

    // Notifications
    getNotifications: (page = 1, limit = 20) => api.request('GET', `/notifications?page=${page}&limit=${limit}`),
    markNotificationAsRead: (id) => api.request('PATCH', `/notifications/${id}/read`),
    markAllNotificationsAsRead: () => api.request('PATCH', `/notifications/read-all`),

    // Watchlist
    getWatchlist: () => api.request('GET', '/watchlist'),
    toggleWatchlist: (auctionId) => api.request('POST', `/watchlist/${auctionId}`),
    checkWatchlistStatus: (auctionId) => api.request('GET', `/watchlist/${auctionId}/status`),

    // Auth
    login: (email, password) => api.request('POST', '/auth/login', { email, password }),
    register: (body) => api.request('POST', '/auth/register', body),

    // Profile
    getMe: () => api.request('GET', '/users/me'),
    updateMe: (body) => api.request('PUT', '/users/me', body),
    getMyBids: () => api.request('GET', '/users/me/bids'),
    getMyWonAuctions: () => api.request('GET', '/users/me/auctions/won'),

    // Users (Admin)
    getUsers: () => api.request('GET', '/users'),
    createUser: (body) => api.request('POST', '/users', body),
    updateUser: (id, body) => api.request('PUT', `/users/${id}`, body),
    deleteUser: (id) => api.request('DELETE', `/users/${id}`),
    lockUser: (id) => api.request('PATCH', `/users/${id}/lock`),
    unlockUser: (id) => api.request('PATCH', `/users/${id}/unlock`),

    // Wallet
    deposit: (amount) => api.request('POST', '/wallet/deposit', { amount }),
    withdraw: (amount) => api.request('POST', '/wallet/withdraw', { amount }),
    getTransactions: () => api.request('GET', '/wallet/transactions'),

    // Auctions
    getAuctions: (filters = {}) => {
        const params = new URLSearchParams();
        if (filters.status) params.append('status', filters.status);
        if (filters.product) params.append('product', filters.product);
        if (filters.seller) params.append('seller', filters.seller);
        if (filters.seller_id) params.append('seller_id', filters.seller_id);
        if (filters.categories && filters.categories.length) params.append('categories', filters.categories.join(','));
        if (filters.page) params.append('page', filters.page);
        if (filters.limit) params.append('limit', filters.limit);
        return api.request('GET', `/auctions?${params.toString()}`);
    },
    getHotAuctions: () => api.request('GET', '/auctions/hot'),
    getAuction: (id) => api.request('GET', `/auctions/${id}`),
    placeBid: (auctionId, amount) => api.request('POST', `/auctions/${auctionId}/bids`, { amount }),
    createAuction: (body) => api.request('POST', '/auctions', body),
    extendAuction: (id, newEndTime) => api.request('PATCH', `/auctions/${id}/extend`, { new_end_time: newEndTime }),
    confirmDelivery: (id) => api.request('PATCH', `/auctions/${id}/confirm-delivery`),
    confirmReceipt: (id) => api.request('PATCH', `/auctions/${id}/confirm-receipt`),
    rejectAuction: (id, reason = '') => api.request('PATCH', `/auctions/${id}/reject`, { reason }),
    
    // Categories
    getCategories: (status = '') => api.request('GET', `/categories${status ? '?status=' + status : ''}`),
    getMyCategories: () => api.request('GET', '/categories/me'),
    createCategory: (body) => api.request('POST', '/categories', body),
    approveCategory: (id, status, reason = '') => api.request('PATCH', `/categories/${id}/status`, { status, reason }),

    // Products
    getProducts: (params = {}) => {
        const query = new URLSearchParams(params).toString();
        return api.request('GET', `/products${query ? '?' + query : ''}`);
    },
    getProduct: (id) => api.request('GET', `/products/${id}`),
    approveProduct: (id, status, reason = '') => api.request('PATCH', `/products/${id}/status`, { status, reason }),
    updateProduct: (id, body) => api.request('PUT', `/products/${id}`, body),
    deleteProduct: (id) => api.request('DELETE', `/products/${id}`),
    deleteProductImage: (productId, imageId) => api.request('DELETE', `/products/${productId}/images/${imageId}`),
    lockProduct: (id, reason = '') => api.request('PATCH', `/products/${id}/lock`, { reason }),
    createProduct: async (formData) => {
        const headers = {};
        if (api._token) headers['Authorization'] = `Bearer ${api._token}`;
        // Note: fetch will automatically set the correct Content-Type with boundary for FormData if we leave it empty
        const res = await fetch(`${API_BASE}/products`, {
            method: 'POST',
            headers,
            body: formData
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Lỗi upload sản phẩm');
        return data;
    },

    logout() {
        this.token = null;
        this.user = null;
    }
};
