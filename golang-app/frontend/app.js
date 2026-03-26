// ========================================
// AuctionHub — Main SPA Application
// ========================================

const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

// ========== UTILITIES ==========
function toast(msg, type = 'info') {
    const el = document.createElement('div');
    el.className = `toast toast-${type}`;
    el.textContent = msg;
    $('#toast-container').appendChild(el);
    setTimeout(() => el.remove(), 3000);
}

function formatMoney(n) {
    return new Intl.NumberFormat('vi-VN').format(n) + ' ₫';
}

function formatDate(d) {
    if (!d) return '—';
    return new Date(d).toLocaleString('vi-VN', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' });
}

function getProductImage(product) {
    if (product && product.images && product.images.length > 0) {
        const primary = product.images.find(img => img.is_primary);
        return primary ? primary.image_url : product.images[0].image_url;
    }
    return `https://ui-avatars.com/api/?name=${encodeURIComponent(product?.name || 'P')}&background=random&size=256`; 
}

function statusBadge(status) {
    const map = { active: 'active', pending: 'pending', ended: 'ended', sold: 'sold', cancelled: 'ended', approved: 'active', rejected: 'ended' };
    const textMap = { active: 'Đang đấu giá', pending: 'Chờ duyệt', ended: 'Đã kết thúc', sold: 'Thành công', cancelled: 'Đã hủy', approved: 'Đã duyệt', rejected: 'Từ chối' };
    return `<span class="badge badge-${map[status] || 'pending'}">${textMap[status] || status}</span>`;
}

function auctionStatusBadge(a) {
    if (!a) return '';
    if (a.status === 'active') return `<span class="badge badge-active">Đang diễn ra</span>`;
    if (a.status === 'ended' || a.status === 'sold') {
        if (!a.winner_id) return `<span class="badge badge-ended">Kết thúc (Không có người mua)</span>`;
        if (a.status === 'sold' || (a.seller_confirmed && a.buyer_confirmed)) return `<span class="badge badge-sold">Giao dịch thành công</span>`;
        if (a.seller_confirmed && !a.buyer_confirmed) return `<span class="badge badge-pending">Chờ khách nhận hàng</span>`;
        if (!a.seller_confirmed && a.buyer_confirmed) return `<span class="badge badge-pending">Chờ người bán gửi hàng</span>`;
        return `<span class="badge badge-ended">Đang chờ xác nhận</span>`;
    }
    return statusBadge(a.status);
}

function roleBadge(role) {
    return `<span class="badge badge-${role}">${role}</span>`;
}

function loading() {
    return '<div class="loading"><div class="spinner"></div></div>';
}

function emptyState(icon, text) {
    return `<div class="empty-state"><div class="empty-icon">${icon}</div><p>${text}</p></div>`;
}

// ========== ROUTER ==========
const pages = {};
let currentPage = '';

const PUBLIC_PAGES = new Set(['login', 'register', 'home', 'auctions', 'auction-room', 'categories']);

function navigate(path, params = null) {
    const [firstSegment] = path.split('/');
    if (!api.isLoggedIn && !PUBLIC_PAGES.has(firstSegment)) {
        toast('Bạn cần đăng nhập để truy cập tính năng này', 'warning');
        path = 'login';
    }
    window.location.hash = path;
    const [page] = path.split('/');
    currentPage = page;
    renderPage(path, params);

    // Update active nav
    $$('.nav-link').forEach(l => l.classList.remove('active'));
    const active = $(`.nav-link[data-page="${page}"]`);
    if (active) active.classList.add('active');
}

function renderPage(path, navParams = null) {
    const [page, ...urlParams] = path.split('/');
    const fn = pages[page];
    if (fn) {
        if (navParams) {
            fn(navParams);
        } else {
            // Handle URL params if any (e.g. auction-detail/13)
            if (urlParams.length > 0) {
                // If it's manage-products/2, wrap it in an object as the page expects
                if (page === 'manage-products') {
                    fn({ page: urlParams[0] });
                } else {
                    fn(...urlParams);
                }
            } else {
                fn();
            }
        }
    }
    else pages['dashboard']?.();
}

// ========== AUTH: LOGIN ==========
pages['login'] = () => {
    $('#sidebar').classList.add('hidden');
    $('#main-content').classList.remove('with-sidebar');
    hideGuestHeader();

    $('#app').innerHTML = `
    <div class="auth-container">
        <div class="auth-card">
            <div class="auth-logo">🔨</div>
            <h1>Đăng nhập</h1>
            <p class="subtitle">Chào mừng bạn đến với AuctionHub</p>
            <form id="login-form">
                <div class="form-group">
                    <label>Email</label>
                    <input type="email" class="form-control" id="login-email" placeholder="email@example.com" required>
                </div>
                <div class="form-group">
                    <label>Mật khẩu</label>
                    <input type="password" class="form-control" id="login-password" placeholder="••••••••" required>
                </div>
                <button type="submit" class="btn btn-primary btn-block" id="login-btn">Đăng nhập</button>
            </form>
            <p class="text-center mt-4 text-sm">
                Chưa có tài khoản? <a class="link" onclick="navigate('register')">Đăng ký ngay</a>
            </p>
        </div>
    </div>`;

    $('#login-form').onsubmit = async (e) => {
        e.preventDefault();
        const btn = $('#login-btn');
        btn.disabled = true;
        btn.textContent = 'Đang xử lý...';
        try {
            const res = await api.login($('#login-email').value, $('#login-password').value);
            api.token = res.data.token;
            api.user = res.data.user;
            toast('Đăng nhập thành công!', 'success');
            initNotifications(); // Initialize real-time notifications
            navigate('dashboard');
        } catch (err) {
            toast(err.message, 'error');
        } finally {
            btn.disabled = false;
            btn.textContent = 'Đăng nhập';
        }
    };
};

// ========== AUTH: REGISTER ==========
pages['register'] = () => {
    $('#sidebar').classList.add('hidden');
    $('#main-content').classList.remove('with-sidebar');
    hideGuestHeader();

    $('#app').innerHTML = `
    <div class="auth-container">
        <div class="auth-card">
            <div class="auth-logo">📝</div>
            <h1>Đăng ký tài khoản</h1>
            <p class="subtitle">Tham gia đấu giá ngay hôm nay</p>
            <form id="register-form">
                <div class="form-group">
                    <label>Họ tên</label>
                    <input type="text" class="form-control" id="reg-name" placeholder="Nguyễn Văn A" required>
                </div>
                <div class="form-group">
                    <label>Email</label>
                    <input type="email" class="form-control" id="reg-email" placeholder="email@example.com" required>
                </div>
                <div class="form-group">
                    <label>Số điện thoại</label>
                    <input type="text" class="form-control" id="reg-phone" placeholder="0901234567" required>
                </div>
                <div class="form-group">
                    <label>Mật khẩu</label>
                    <input type="password" class="form-control" id="reg-password" placeholder="••••••••" required>
                </div>
                <div class="form-group">
                    <label>Vai trò</label>
                    <select class="form-control" id="reg-role">
                        <option value="bidder">Người mua (Bidder)</option>
                        <option value="seller">Người bán (Seller)</option>
                    </select>
                </div>
                <button type="submit" class="btn btn-primary btn-block" id="reg-btn">Đăng ký</button>
            </form>
            <p class="text-center mt-4 text-sm">
                Đã có tài khoản? <a class="link" onclick="navigate('login')">Đăng nhập</a>
            </p>
        </div>
    </div>`;

    $('#register-form').onsubmit = async (e) => {
        e.preventDefault();
        const btn = $('#reg-btn');
        btn.disabled = true;
        try {
            await api.register({
                full_name: $('#reg-name').value,
                email: $('#reg-email').value,
                phone_number: $('#reg-phone').value,
                password: $('#reg-password').value,
                role: $('#reg-role').value
            });
            toast('Đăng ký thành công! Hãy đăng nhập.', 'success');
            navigate('login');
        } catch (err) {
            toast(err.message, 'error');
        } finally {
            btn.disabled = false;
        }
    };
};

// ========== SHOW SIDEBAR ==========
function showSidebar() {
    hideGuestHeader();
    $('#sidebar').classList.remove('hidden');
    $('#main-content').classList.add('with-sidebar');
    const u = api.user;
    $('#user-info').innerHTML = `<div class="user-name">${u.full_name}</div><div class="user-role">${roleBadge(u.role)} ${u.email}</div>`;
    
    // Toggle Admin Links
    const isAdmin = api.isAdmin;
    const isSeller = u.role === 'seller';

    if (isAdmin) {
        $('#admin-link').classList.remove('hidden');
        $('#admin-category-link').classList.remove('hidden');
    } else {
        $('#admin-link').classList.add('hidden');
        $('#admin-category-link').classList.add('hidden');
    }

    // Toggle Seller Links
    if (isSeller || isAdmin) {
        $('#seller-product-link').classList.remove('hidden');
    } else {
        $('#seller-product-link').classList.add('hidden');
    }

    // New: Hide Wallet and My Bids for Admin
    if (isAdmin) {
        $(`.nav-link[data-page="wallet"]`)?.parentElement.classList.add('hidden');
        $(`.nav-link[data-page="my-bids"]`)?.parentElement.classList.add('hidden');
        $(`.nav-link[data-page="won"]`)?.parentElement.classList.add('hidden');
    } else {
        $(`.nav-link[data-page="wallet"]`)?.parentElement.classList.remove('hidden');
        $(`.nav-link[data-page="my-bids"]`)?.parentElement.classList.remove('hidden');
        $(`.nav-link[data-page="won"]`)?.parentElement.classList.remove('hidden');
    }
}

// ========== GUEST HEADER HELPERS ==========
function showGuestHeader() {
    $('#sidebar').classList.add('hidden');
    $('#main-content').classList.remove('with-sidebar');
    const gh = $('#guest-header');
    if (gh) gh.classList.remove('hidden');
    $('#main-content').classList.add('with-guest-header');
}
function hideGuestHeader() {
    const gh = $('#guest-header');
    if (gh) gh.classList.add('hidden');
    $('#main-content').classList.remove('with-guest-header');
}

// ========== HOME (PUBLIC LANDING PAGE) ==========
pages['home'] = async () => {
    if (api.isLoggedIn) { navigate('dashboard'); return; }
    showGuestHeader();
    $('#app').innerHTML = `
    <div class="home-hero">
        <div class="hero-badge">🔨 AuctionHub</div>
        <h1 class="hero-title">Sàn đấu giá trực tuyến<br><span class="hero-highlight">Minh bạch · Thời gian thực</span></h1>
        <p class="hero-sub">Hàng ngàn sản phẩm độc đáo, phiên đấu giá mới mỗi ngày.<br>Đăng ký miễn phí và bắt đầu ngay hôm nay.</p>
        <div class="hero-actions">
            <button class="btn btn-primary btn-lg" onclick="navigate('auctions')">🔍 Khám phá đấu giá</button>
            <button class="btn btn-outline btn-lg" onclick="navigate('register')">🚀 Tham gia miễn phí</button>
        </div>
    </div>
    <div id="home-content">${loading()}</div>`;
    try {
        const [hotRes, catRes] = await Promise.all([
            api.getHotAuctions(),
            api.getCategories('active')
        ]);
        const hotList = (hotRes.data || []).slice(0, 6);
        const categories = (catRes.data || []).slice(0, 8);
        const hotHTML = hotList.length ? `
        <div class="home-section">
            <div class="section-header">
                <h2>🔥 Phiên đấu giá HOT</h2>
                <button class="btn btn-outline btn-sm" onclick="navigate('auctions')">Xem tất cả →</button>
            </div>
            <div class="auction-grid">
                ${hotList.map(a => `
                <div class="auction-card" onclick="navigate('auction-room/${a.id}')" style="cursor:pointer">
                    <div class="card-image-wrapper">
                        <img src="${getProductImage(a.product)}" class="auction-card-img" alt="${a.product?.name}">
                    </div>
                    <div class="flex-between mt-2">
                        <span class="auction-name">${a.product?.name || 'Phiên #' + a.id}</span>
                        ${statusBadge(a.status)}
                    </div>
                    <div class="auction-price mt-1">${formatMoney(a.current_price || a.start_price)}</div>
                    <div class="flex-between mt-2">
                        <span class="bid-count">🔨 ${a.bid_count || 0} lượt bid</span>
                        <span class="text-xs text-muted">${formatDate(a.end_time)}</span>
                    </div>
                </div>`).join('')}
            </div>
        </div>` : '';
        const catHTML = categories.length ? `
        <div class="home-section">
            <div class="section-header">
                <h2>📁 Danh mục sản phẩm</h2>
                <button class="btn btn-outline btn-sm" onclick="navigate('categories')">Xem tất cả →</button>
            </div>
            <div class="category-grid-home">
                ${categories.map(c => `
                <div class="cat-card-home" onclick="window.applyCategoryFilterAndNavigate(${c.id})">
                    <div class="cat-icon">📦</div>
                    <div class="cat-name">${c.name}</div>
                    <div class="cat-desc">${c.description || ''}</div>
                </div>`).join('')}
            </div>
        </div>` : '';
        const ctaHTML = `
        <div class="home-cta-block">
            <h2>Bạn chưa có tài khoản?</h2>
            <p>Đăng ký miễn phí để đặt giá, theo dõi phiên và nhận thông báo thời gian thực.</p>
            <div class="hero-actions">
                <button class="btn btn-primary btn-lg" onclick="navigate('register')">Đăng ký ngay</button>
                <button class="btn btn-outline btn-lg" onclick="navigate('login')">Đăng nhập</button>
            </div>
        </div>`;
        $('#home-content').innerHTML = hotHTML + catHTML + ctaHTML;
    } catch (err) {
        $('#home-content').innerHTML = emptyState('⚠️', 'Không thể tải dữ liệu. Vui lòng thử lại.');
    }
};

// ========== DASHBOARD ===========
pages['dashboard'] = async () => {
    showSidebar();
    $('#app').innerHTML = `<div class="page-header"><h1>👋 Xin chào, ${api.user.full_name}</h1><p>Tổng quan tài khoản của bạn</p></div>${loading()}`;

    try {
        const [profile, bids, won, hot] = await Promise.all([
            api.getMe(), api.getMyBids(), api.getMyWonAuctions(), api.getHotAuctions()
        ]);
        const u = profile.data;
        const w = u.wallet || { balance: 0, frozen_balance: 0 };
        const bidList = bids.data || [];
        const wonList = won.data || [];
        const hotList = hot.data || [];

        $('#app').innerHTML = `
        <div class="page-header">
            <h1>👋 Xin chào, ${u.full_name}</h1>
            <p>Tổng quan tài khoản của bạn</p>
        </div>
        <div class="stat-grid">
            <div class="stat-card">
                <div class="stat-label">Số dư khả dụng</div>
                <div class="stat-value text-success">${formatMoney(w.balance - w.frozen_balance)}</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Đang đóng băng</div>
                <div class="stat-value text-warning">${formatMoney(w.frozen_balance)}</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Tổng lượt bid</div>
                <div class="stat-value text-accent">${bidList.length}</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Đã thắng</div>
                <div class="stat-value text-danger">${wonList.length} phiên</div>
            </div>
        </div>

        ${hotList.length ? `
        <div class="mb-4">
            <h2 class="mb-3">🔥 Đấu giá HOT nhất</h2>
            <div class="auction-grid">
                ${hotList.slice(0, 3).map(a => `
                <div class="auction-card" onclick="confirmJoinAuction(${a.id}, '${a.product?.name}')">
                    <img src="${getProductImage(a.product)}" class="auction-card-img" alt="${a.product?.name}">
                    <div class="flex-between">
                        <span class="auction-name">${a.product?.name}</span>
                        ${statusBadge(a.status)}
                    </div>
                    <div class="auction-price">${formatMoney(a.current_price)}</div>
                    <div class="flex-between text-xs text-muted">
                        <span>🔨 ${a.bid_count} lượt bid</span>
                        ${a.bids && a.bids.length > 0 ? `<span class="card-high-bidder">👑 ${a.bids[0].user?.full_name}</span>` : ''}
                    </div>
                </div>`).join('')}
            </div>
        </div>` : ''}

        <div class="card">
            <div class="card-header"><h2>📋 Bid gần đây</h2></div>
            <div class="card-body">
                ${bidList.length ? `
                <div class="table-wrapper"><table>
                    <thead><tr><th>Phiên</th><th>Giá Bid</th><th>Thời gian</th><th>Trạng thái</th></tr></thead>
                    <tbody>${bidList.slice(0, 5).map(b => `
                        <tr>
                            <td>${b.auction?.product?.name || `Phiên #${b.auction_id}`}</td>
                            <td><strong>${formatMoney(b.amount)}</strong></td>
                            <td class="text-muted">${formatDate(b.bid_time)}</td>
                            <td>${statusBadge(b.auction?.status || 'active')}</td>
                        </tr>`).join('')}
                    </tbody>
                </table></div>` : emptyState('📋', 'Bạn chưa có lượt bid nào')}
            </div>
        </div>`;
    } catch (err) {
        toast(err.message, 'error');
    }
};

// ========== AUCTIONS ==========
let auctionFilters = {
    page: 1,
    limit: 12,
    status: '',
    product: '',
    seller: '',
    seller_id: 0,
    categories: []
};

pages['auctions'] = async () => {
    if (api.isLoggedIn) showSidebar(); else showGuestHeader();
    
    const categoriesRes = await api.getCategories('active');
    const allCategories = categoriesRes.data || [];

    const filterHTML = `
    <div class="search-bar-luxury">
        <div class="search-input-wrapper">
            <i class="search-icon">🔍</i>
            <input type="text" id="f-product" class="form-control" placeholder="Tìm tên sản phẩm hoặc từ khóa..." value="${auctionFilters.product}">
        </div>
        <button class="btn btn-primary" style="border-radius:28px; padding:0 35px;" onclick="applyAuctionFilters()">Tìm kiếm</button>
    </div>

    <div class="filter-section">
        <div class="filter-title">
            <span>🏷️ Lọc theo danh mục</span>
            <span class="text-xs text-muted" style="font-weight:400;">(Chọn nhiều danh mục để thu hẹp kết quả)</span>
        </div>
        <div class="category-tags">
            ${allCategories.map(c => `
                <div class="category-tag ${auctionFilters.categories.includes(c.id) ? 'active' : ''}" 
                     onclick="toggleCategoryFilter(${c.id})">
                    ${c.name}
                </div>
            `).join('')}
        </div>
        ${auctionFilters.categories.length > 0 ? `
            <div class="active-filters">
                ${auctionFilters.categories.map(id => {
                    const cat = allCategories.find(c => c.id === id);
                    return `<div class="filter-chip">
                        ${cat ? cat.name : id}
                        <b class="remove-filter" onclick="toggleCategoryFilter(${id})">✕</b>
                    </div>`;
                }).join('')}
                <button class="btn btn-link text-xs" style="color:var(--danger);" onclick="clearCategoryFilters()">Xóa tất cả</button>
            </div>
        ` : ''}

        <div class="grid mt-6 pt-4 border-t border-dashed border-border" style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
            <div class="form-group">
                <label class="text-xs text-muted mb-1 block">Người chủ trì</label>
                <input type="text" id="f-seller" class="form-control" placeholder="Tên seller..." value="${auctionFilters.seller}">
            </div>
            <div class="form-group">
                <label class="text-xs text-muted mb-1 block">Trạng thái phiên</label>
                <select id="f-status" class="form-control">
                    <option value="">Tất cả trạng thái</option>
                    <option value="active" ${auctionFilters.status === 'active' ? 'selected' : ''}>Đang diễn ra</option>
                    <option value="pending" ${auctionFilters.status === 'pending' ? 'selected' : ''}>Sắp diễn ra</option>
                    <option value="ended" ${auctionFilters.status === 'ended' ? 'selected' : ''}>Đã kết thúc</option>
                </select>
            </div>
        </div>
        ${api.user?.role === 'seller' ? `
            <div class="mt-4 pt-4 border-t border-border">
                <label class="flex items-center gap-2 cursor-pointer" style="width:fit-content;">
                    <input type="checkbox" id="f-my-auctions" ${auctionFilters.seller_id ? 'checked' : ''} onchange="toggleMyAuctions(this.checked)">
                    <span class="text-sm font-semibold">💎 Chỉ xem phiên của tôi</span>
                </label>
            </div>
        ` : ''}
    </div>`;

    $('#app').innerHTML = `
        <div class="page-header">
            <div>
                <h1>📦 Khám phá đấu giá</h1>
                <p>Tìm kiếm và tham dự các phiên đấu giá công khai</p>
            </div>
        </div>
        ${filterHTML}
        <div id="auction-list-container">${loading()}</div>`;

    try {
        const res = await api.getAuctions(auctionFilters);
        const { auctions, total_count, page, limit } = res.data;
        
        // Get user's watchlist to highlight followed auctions
        let watchlistIds = [];
        if (api.isLoggedIn) {
            try {
                const wRes = await api.getWatchlist();
                watchlistIds = (wRes.data || []).map(item => item.auction_id);
            } catch (e) { console.error('Watchlist fetch error:', e); }
        }

        if (!auctions.length) {
            $('#auction-list-container').innerHTML = emptyState('🔎', 'Không tìm thấy phiên đấu giá nào phù hợp');
            return;
        }

        const gridHTML = auctions.map(a => {
            const isWatching = watchlistIds.includes(a.id);
            return `
            <div class="auction-card" onclick="confirmJoinAuction(${a.id}, '${a.product?.name || `Phiên #${a.id}`}')">
                <div class="card-image-wrapper">
                    <img src="${getProductImage(a.product)}" class="auction-card-img" alt="${a.product?.name || `Phiên #${a.id}`}">
                    <button class="btn-watch ${isWatching ? 'active' : ''}" onclick="event.stopPropagation(); window.toggleWatch(${a.id}, this)">
                        ${isWatching ? '❤️' : '🤍'}
                    </button>
                </div>
                <div class="flex-between">
                    <span class="auction-name">${a.product?.name || `Phiên #${a.id}`}</span>
                    ${statusBadge(a.status)}
                </div>
                <div class="auction-meta">
                    <span>Người chủ trì: <strong>${a.product?.seller?.full_name || 'Hệ thống'}</strong></span>
                </div>
                <div class="auction-meta"><span>Giá khởi điểm</span><span>${formatMoney(a.start_price)}</span></div>
                <div class="auction-price">${formatMoney(a.current_price)}</div>
                <div class="flex-between mt-2">
                    <span class="bid-count">🔨 ${a.bid_count || 0} lượt bid</span>
                    ${a.bids && a.bids.length > 0 ? `<span class="card-high-bidder">👑 ${a.bids[0].user?.full_name}</span>` : ''}
                </div>
                <div class="text-right mt-1">
                    <span class="text-xs text-muted">Kết thúc: ${formatDate(a.end_time)}</span>
                </div>
            </div>`;
        }).join('');

        const totalPages = Math.ceil(total_count / limit);
        const paginationHTML = totalPages > 1 ? `
        <div class="pagination mt-4">
            <button class="btn btn-outline btn-sm" ${page === 1 ? 'disabled' : ''} onclick="changeAuctionPage(${page - 1})">Trước</button>
            <span class="text-sm">Trang ${page} / ${totalPages}</span>
            <button class="btn btn-outline btn-sm" ${page === totalPages ? 'disabled' : ''} onclick="changeAuctionPage(${page + 1})">Sau</button>
        </div>` : '';

        $('#auction-list-container').innerHTML = `<div class="auction-grid">${gridHTML}</div>` + paginationHTML;
    } catch (err) {
        toast(err.message, 'error');
    }
};

window.toggleCategoryFilter = (id) => {
    const idx = auctionFilters.categories.indexOf(id);
    if (idx > -1) {
        auctionFilters.categories.splice(idx, 1);
    } else {
        auctionFilters.categories.push(id);
    }
    applyAuctionFilters();
};

window.clearCategoryFilters = () => {
    auctionFilters.categories = [];
    applyAuctionFilters();
};

window.applyAuctionFilters = () => {
    auctionFilters.product = $('#f-product').value;
    const sellerInput = $('#f-seller');
    auctionFilters.seller = sellerInput ? sellerInput.value : '';
    auctionFilters.status = $('#f-status').value;
    auctionFilters.page = 1; // Reset to first page
    navigate('auctions');
};

window.toggleMyAuctions = (checked) => {
    auctionFilters.seller_id = checked ? api.user.id : 0;
    // If filtering by my auctions, clear the seller name filter to avoid confusion
    if (checked) {
        auctionFilters.seller = '';
        const sellerInput = $('#f-seller');
        if (sellerInput) sellerInput.value = '';
    }
    applyAuctionFilters();
};

window.changeAuctionPage = (p) => {
    auctionFilters.page = p;
    navigate('auctions');
};

// ========== AUCTION DETAIL (Modal) ==========
let countdownInterval = null;

function updateCountdown(endTime, containerId) {
    const target = new Date(endTime).getTime();
    const el = document.getElementById(containerId);
    if (!el) return;

    if (countdownInterval) clearInterval(countdownInterval);

    countdownInterval = setInterval(() => {
        const now = new Date().getTime();
        const diff = target - now;

        if (diff <= 0) {
            el.innerHTML = `<div class="badge badge-ended">Phiên đấu giá đã kết thúc</div>`;
            clearInterval(countdownInterval);
            return;
        }

        const d = Math.floor(diff / (1000 * 60 * 60 * 24));
        const h = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
        const m = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
        const s = Math.floor((diff % (1000 * 60)) / 1000);

        const isUrgent = diff < 60000; // Less than 1 minute

        el.innerHTML = `
            <div class="countdown-container ${isUrgent ? 'timer-pulse' : ''}">
                <div class="timer-segment"><span class="timer-value">${h}</span><span class="timer-label">Giờ</span></div>
                <div class="timer-segment"><span class="timer-value">${m}</span><span class="timer-label">Phút</span></div>
                <div class="timer-segment"><span class="timer-value">${s}</span><span class="timer-label">Giây</span></div>
            </div>
        `;
    }, 1000);
}

window.confirmJoinAuction = async (id, name) => {
    if (!api.isLoggedIn) {
        navigate(`auction-room/${id}`);
        return;
    }
    const modal = document.createElement('div');
    modal.className = 'modal-overlay';
    modal.innerHTML = `
    <div class="modal">
        <div class="room-confirmation">
            <span class="room-icon">🚪</span>
            <h2>Vào phòng đấu giá?</h2>
            <p class="text-secondary mb-4">Bạn có muốn tham gia phòng đấu giá cho sản phẩm <strong>${name}</strong> không?</p>
            <div class="modal-actions" style="justify-content: center;">
                <button class="btn btn-outline" onclick="this.closest('.modal-overlay').remove()">Hủy</button>
                <button class="btn btn-primary" id="confirm-join-btn">🚀 Tham gia ngay</button>
            </div>
        </div>
    </div>`;
    document.body.appendChild(modal);
    $('#confirm-join-btn').onclick = () => {
        modal.remove();
        navigate(`auction-room/${id}`);
    };
};

pages['auction-room'] = async (id) => {
    if (!id) return navigate('auctions');
    if (api.isLoggedIn) showSidebar(); else showGuestHeader();
    $('#app').innerHTML = loading();

    try {
        const res = await api.getAuction(id);
        const a = res.data;
        const isSeller = api.user?.id === a.product?.seller_id;

        $('#app').innerHTML = `
        <div class="page-header">
            <div class="flex-between">
                <div>
                    <h1>🏁 Phòng đấu giá: ${a.product?.name || 'Đang tải...'}</h1>
                    <p>Phiên #${a.id} • Người chủ trì: <strong>${a.product?.seller?.full_name || 'Hệ thống'}</strong></p>
                </div>
                <button class="btn btn-outline" onclick="navigate('auctions')">⬅ Quay lại</button>
            </div>
        </div>

        <div class="room-container room-layout-v2">
            <div class="room-top-content">
                <div class="room-gallery">
                    ${(() => {
                        const images = a.product?.images || [];
                        if (images.length === 0) {
                            return `<div class="room-slider-container"><div class="room-slide active"><img src="${getProductImage(a.product)}" alt="${a.product?.name}"></div></div>`;
                        }
                        return `
                        <div class="room-slider-container">
                            ${images.map((img, idx) => `
                                <div class="room-slide ${idx === 0 ? 'active' : ''}">
                                    <img src="${img.image_url}" alt="${a.product.name}">
                                </div>
                            `).join('')}
                            ${images.length > 1 ? `
                                <button class="slider-nav slider-prev" onclick="moveSlide(-1)">❮</button>
                                <button class="slider-nav slider-next" onclick="moveSlide(1)">❯</button>
                                <div class="slider-dots">
                                    ${images.map((_, idx) => `<div class="slider-dot ${idx === 0 ? 'active' : ''}" onclick="goToSlide(${idx})"></div>`).join('')}
                                </div>
                            ` : ''}
                        </div>`;
                    })()}
                </div>

                <div class="room-action-panel">
                    <div class="room-hero">
                        <div id="room-countdown"></div>
                        <div class="room-price-box">
                            <span class="room-price-label">Giá hiện tại</span>
                            <div class="room-current-price" id="room-price-display">${formatMoney(a.current_price)}</div>
                            <div id="room-high-bidder">
                                ${a.bids && a.bids.length > 0 ? `<div class="room-high-bidder">👑 <strong>${a.bids[0].user?.full_name}</strong></div>` : '<div class="room-high-bidder text-sm">🍀 Hãy đặt giá đầu tiên</div>'}
                            </div>
                        </div>
                        <div class="flex-between mt-3">
                            ${statusBadge(a.status)}
                            <span class="text-xs text-muted" id="room-bid-count">🔨 ${a.bid_count || 0} lượt bid</span>
                        </div>
                        ${['banned', 'cancelled', 'rejected'].includes(a.status) && a.rejection_reason ? `
                            <div class="alert alert-danger text-xs mt-3">
                                <strong>Lý do:</strong> ${a.rejection_reason}
                            </div>
                        ` : ''}
                    </div>

                    <div class="stat-grid-compact">
                        <div class="stat-item"><span>Khởi điểm:</span> <strong>${formatMoney(a.start_price)}</strong></div>
                        <div class="stat-item"><span>Bước giá:</span> <strong>${formatMoney(a.step_price)}</strong></div>
                        ${a.buy_now_price ? `<div class="stat-item"><span>Mua ngay:</span> <strong class="text-accent">${formatMoney(a.buy_now_price)}</strong></div>` : ''}
                    </div>

                    <div class="room-bidding-area">
                        ${!api.isLoggedIn ? `
                            <div class="guest-bid-cta">
                                <p class="text-secondary mb-3" style="font-size:.95rem;">Đăng nhập để tham gia đặt giá</p>
                                <div class="flex gap-3" style="justify-content:center;">
                                    <button class="btn btn-primary" onclick="navigate('login')">Đăng nhập</button>
                                    <button class="btn btn-outline" onclick="navigate('register')">Đăng ký</button>
                                </div>
                            </div>
                        ` : (api.user?.role === 'admin' ? `
                            <div class="alert alert-warning text-xs mb-2">🛡️ Admin không được phép đặt giá.</div>
                            <button class="btn btn-danger btn-sm btn-block" onclick="rejectAuction(${a.id})">⛔ Ban phiên đấu giá này</button>
                        ` : (isSeller ? `
                            <div class="alert alert-info text-xs mb-3">🌟 Bạn là người chủ trì phiên này.</div>
                            <button class="btn btn-danger btn-outline btn-sm btn-block" onclick="rejectAuction(${a.id})">❌ Hủy phiên của tôi</button>
                        ` : (a.status === 'active' && api.user?.role === 'bidder' ? `
                            <div class="bid-controls-v2">
                                <label class="text-xs text-muted mb-2 block">Đặt giá của bạn (≥ ${formatMoney((a.current_price || a.start_price) + (a.step_price || 0))})</label>
                                <div class="flex gap-2">
                                    <input type="number" class="form-control" id="bid-amount" style="font-weight:700; flex: 1;" placeholder="Số tiền..." min="${(a.current_price || a.start_price) + (a.step_price || 0)}">
                                    <button class="btn btn-primary" id="bid-btn" style="white-space:nowrap;" onclick="submitRoomBid(${a.id})">🔨 Bid</button>
                                </div>
                                ${a.buy_now_price ? `
                                    <button class="btn btn-accent btn-outline btn-block mt-3" onclick="buyNow(${a.id}, ${a.buy_now_price})">
                                        ⚡ Mua ngay: ${formatMoney(a.buy_now_price)}
                                    </button>
                                ` : ''}
                            </div>
                        ` : `<div class="text-center text-muted text-sm mt-4">Phiên đấu giá ${a.status === 'ended' ? 'đã kết thúc' : 'hiện không nhận bid'}</div>`)))}
                    </div>
                </div>
            </div>

            <div class="room-bottom-content">
                <div class="room-bid-history-v2">
                    <div class="history-header">
                        <h3>🕒 Lịch sử phòng đấu giá</h3>
                    </div>
                    <div class="history-grid" id="room-history-list">
                        ${a.bids && a.bids.length > 0 ? a.bids.map(b => `
                            <div class="history-card">
                                <div class="history-user">${b.user?.full_name || 'Người tham gia'}</div>
                                <div class="flex-between">
                                    <span class="history-price">${formatMoney(b.amount)}</span>
                                    <span class="history-time">${formatDate(b.bid_time)}</span>
                                </div>
                            </div>
                        `).join('') : '<div class="empty-state">Chưa có lượt đặt giá nào.</div>'}
                    </div>
                </div>
            </div>
        </div>`;

        updateCountdown(a.end_time, 'room-countdown');
        connectRoomWS(id);
        currentSlide = 0; // Reset slider
    } catch (err) {
        toast(err.message, 'error');
        navigate('auctions');
    }
};

let currentSlide = 0;
window.moveSlide = (n) => {
    const slides = document.querySelectorAll('.room-slide');
    if (!slides.length) return;
    currentSlide = (currentSlide + n + slides.length) % slides.length;
    showSlide(currentSlide);
};

window.goToSlide = (n) => {
    currentSlide = n;
    showSlide(currentSlide);
};

function showSlide(n) {
    const slides = document.querySelectorAll('.room-slide');
    const dots = document.querySelectorAll('.slider-dot');
    slides.forEach((s, idx) => s.classList.toggle('active', idx === n));
    dots.forEach((d, idx) => d.classList.toggle('active', idx === n));
}

async function buyNow(id, price) {
    if (!confirm(`Bạn có chắc chắn muốn mua ngay sản phẩm này với giá ${formatMoney(price)}?`)) return;
    try {
        await api.placeBid(id, price);
        toast('Chúc mừng! Bạn đã mua thành công sản phẩm này! 🎉', 'success');
    } catch (err) {
        toast(err.message, 'error');
    }
}

async function submitRoomBid(id) {
    const amount = parseFloat($('#bid-amount')?.value);
    if (!amount || amount <= 0) return toast('Vui lòng nhập số tiền hợp lệ', 'error');
    const btn = $('#bid-btn');
    btn.disabled = true;
    try {
        await api.placeBid(id, amount);
        toast('Đặt giá thành công! 🎉', 'success');
        $('#bid-amount').value = '';
    } catch (err) {
        toast(err.message, 'error');
    } finally {
        if (btn) btn.disabled = false;
    }
}

function connectRoomWS(id) {
    if (auctionWS) auctionWS.close();
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    auctionWS = new WebSocket(`${protocol}//${host}/api/v1/ws/auctions/${id}`);

    auctionWS.onmessage = (e) => {
        const data = JSON.parse(e.data);
        if (data.type === 'bid_update') {
            const priceEl = $('#room-price-display');
            const countEl = $('#room-bid-count');
            const listEl = $('#room-history-list');

            if (priceEl) {
                priceEl.style.animation = 'none';
                priceEl.offsetHeight; // trigger reflow
                priceEl.style.animation = 'fadeIn 0.4s ease';
                priceEl.textContent = formatMoney(data.current_price);
            }
            if (countEl) countEl.textContent = `🔨 ${data.bid_count} lượt đặt giá`;
            
            const highBidderEl = $('#room-high-bidder');
            if (highBidderEl) {
                highBidderEl.innerHTML = `👑 Đang giữ giá: <strong>${data.bidder_name}</strong>`;
            }

            // Real-time Status Update
            if (data.auction && (data.auction.status === 'sold' || data.auction.status === 'ended')) {
                const badgeContainer = $('.inline-flex');
                if (badgeContainer) {
                    badgeContainer.innerHTML = `
                        ${statusBadge(data.auction.status)}
                        <span class="text-muted">🔨 ${data.bid_count} lượt đặt giá</span>
                    `;
                }
                const actionsContainer = $('.room-actions');
                if (actionsContainer) {
                    actionsContainer.innerHTML = `<div class="alert alert-info" style="text-align:center; padding:16px; border:1px solid var(--info); border-radius:var(--radius);">
                        🏁 Phiên đấu giá đã kết thúc (${data.auction.status === 'sold' ? 'Thành công' : 'Không có người mua'}).
                    </div>`;
                }
                if (countdownInterval) clearInterval(countdownInterval);
                const countdownEl = $('#room-countdown');
                if (countdownEl) countdownEl.innerHTML = `<div class="badge badge-ended">Phiên đấu giá đã kết thúc</div>`;
                
                toast('Phiên đấu giá vừa kết thúc! 🏁', 'success');
            } else {
                toast('Giá vừa được cập nhật! 🔥', 'info');
            }

            if (listEl) {
                if (listEl.querySelector('.empty-state')) listEl.innerHTML = '';
                const item = document.createElement('div');
                item.className = 'history-item';
                item.innerHTML = `
                    <div class="flex-between">
                        <span class="history-user">${data.bidder_name}</span>
                        <span class="history-time">${formatDate(new Date())}</span>
                    </div>
                    <div class="history-price">${formatMoney(data.current_price)}</div>
                `;
                listEl.prepend(item);
            }
        }
    };

    auctionWS.onerror = () => console.error('WS Error');
    auctionWS.onclose = () => console.log('WS Closed');
}

let auctionWS = null;

// ========== WALLET ==========
pages['wallet'] = async () => {
    showSidebar();
    $('#app').innerHTML = `<div class="page-header"><h1>💰 Ví tiền</h1><p>Nạp, rút tiền và xem lịch sử giao dịch</p></div>${loading()}`;

    try {
        const [profile, txRes] = await Promise.all([api.getMe(), api.getTransactions()]);
        const w = profile.data.wallet || { balance: 0, frozen_balance: 0 };
        const txs = txRes.data || [];

        $('#app').innerHTML = `
        <div class="page-header"><h1>💰 Ví tiền</h1><p>Nạp, rút tiền và xem lịch sử giao dịch</p></div>
        <div class="stat-grid">
            <div class="stat-card"><div class="stat-label">Tổng số dư</div><div class="stat-value text-success">${formatMoney(w.balance)}</div></div>
            <div class="stat-card"><div class="stat-label">Đang đóng băng</div><div class="stat-value text-warning">${formatMoney(w.frozen_balance)}</div></div>
            <div class="stat-card"><div class="stat-label">Khả dụng</div><div class="stat-value text-accent">${formatMoney(w.balance - w.frozen_balance)}</div></div>
        </div>
        <div class="wallet-actions">
            <div class="wallet-action-card">
                <h3>💳 Nạp tiền</h3>
                <div class="form-group"><input type="number" class="form-control" id="deposit-amount" placeholder="Nhập số tiền nạp" min="1000"></div>
                <button class="btn btn-success btn-block" onclick="doDeposit()">Nạp tiền</button>
            </div>
            <div class="wallet-action-card">
                <h3>🏧 Rút tiền</h3>
                <div class="form-group"><input type="number" class="form-control" id="withdraw-amount" placeholder="Tối đa: ${formatMoney(w.balance - w.frozen_balance)}" max="${w.balance - w.frozen_balance}"></div>
                <button class="btn btn-warning btn-block" onclick="doWithdraw()">Rút tiền</button>
            </div>
        </div>
        <div class="card">
            <div class="card-header"><h2>📜 Lịch sử giao dịch</h2></div>
            <div class="card-body">
                ${txs.length ? `<div class="table-wrapper"><table>
                    <thead><tr><th>Loại</th><th>Số tiền</th><th>Mô tả</th><th>Thời gian</th></tr></thead>
                    <tbody>${txs.map(t => `
                        <tr>
                            <td>${typeBadge(t.type)}</td>
                            <td style="font-weight:600;color:${t.amount >= 0 ? 'var(--success)' : 'var(--danger)'}">${t.amount >= 0 ? '+' : ''}${formatMoney(t.amount)}</td>
                            <td class="text-muted">${t.description || '—'}</td>
                            <td class="text-muted">${formatDate(t.created_at)}</td>
                        </tr>`).join('')}
                    </tbody>
                </table></div>` : emptyState('📜', 'Chưa có giao dịch nào')}
            </div>
        </div>`;
    } catch (err) {
        toast(err.message, 'error');
    }
};

function typeBadge(type) {
    const map = { deposit: ['badge-active', '💳 Nạp'], withdraw: ['badge-ended', '🏧 Rút'], hold: ['badge-pending', '🔒 Hold'], refund: ['badge-sold', '↩️ Hoàn'], payment: ['badge-admin', '💸 TT'] };
    const [cls, label] = map[type] || ['badge-pending', type];
    return `<span class="badge ${cls}">${label}</span>`;
}

async function doDeposit() {
    const amount = parseFloat($('#deposit-amount')?.value);
    if (!amount || amount <= 0) return toast('Nhập số tiền hợp lệ', 'error');
    try { await api.deposit(amount); toast(`Nạp thành công ${formatMoney(amount)}`, 'success'); navigate('wallet'); }
    catch (err) { toast(err.message, 'error'); }
}

async function doWithdraw() {
    const amount = parseFloat($('#withdraw-amount')?.value);
    if (!amount || amount <= 0) return toast('Nhập số tiền hợp lệ', 'error');
    try { await api.withdraw(amount); toast(`Rút thành công ${formatMoney(amount)}`, 'success'); navigate('wallet'); }
    catch (err) { toast(err.message, 'error'); }
}

// ========== MY BIDS ==========
pages['my-bids'] = async () => {
    showSidebar();
    $('#app').innerHTML = `<div class="page-header"><h1>📋 Lịch sử Bid</h1><p>Tất cả lượt đặt giá của bạn</p></div>${loading()}`;

    try {
        const res = await api.getMyBids();
        const list = res.data || [];
        const header = `<div class="page-header"><h1>📋 Lịch sử Bid</h1><p>Tất cả lượt đặt giá của bạn (${list.length} lượt)</p></div>`;

        if (!list.length) { $('#app').innerHTML = header + emptyState('📋', 'Bạn chưa đặt giá phiên nào'); return; }
        $('#app').innerHTML = header + `
        <div class="card"><div class="card-body"><div class="table-wrapper"><table>
            <thead><tr><th>Phiên đấu giá</th><th>Giá bạn bid</th><th>Giá hiện tại</th><th>Trạng thái</th><th>Thời gian</th></tr></thead>
            <tbody>${list.map(b => `
                <tr>
                    <td><a class="link" onclick="navigate('auction-detail/${b.auction_id}')">${b.auction?.product?.name || `#${b.auction_id}`}</a></td>
                    <td><strong>${formatMoney(b.amount)}</strong></td>
                    <td>${formatMoney(b.auction?.current_price || 0)}</td>
                    <td>${auctionStatusBadge(b.auction)}</td>
                    <td class="text-muted">${formatDate(b.bid_time)}</td>
                </tr>`).join('')}
            </tbody>
        </table></div></div></div>`;
    } catch (err) {
        toast(err.message, 'error');
    }
};

// ========== WON AUCTIONS ==========
pages['won'] = async () => {
    showSidebar();
    $('#app').innerHTML = `<div class="page-header"><h1>🏆 Đấu giá đã thắng</h1><p>Các phiên bạn đã chiến thắng</p></div>${loading()}`;

    try {
        const res = await api.getMyWonAuctions();
        const list = res.data || [];
        const header = `<div class="page-header"><h1>🏆 Đấu giá đã thắng</h1><p>Bạn đã thắng ${list.length} phiên đấu giá</p></div>`;

        if (!list.length) { $('#app').innerHTML = header + emptyState('🏆', 'Bạn chưa thắng phiên đấu giá nào'); return; }
        $('#app').innerHTML = header + `
        <div class="card"><div class="card-body"><div class="table-wrapper"><table>
            <thead><tr><th>Sản phẩm</th><th>Người bán</th><th>Giá thắng</th><th>Trạng thái</th><th>Kết thúc</th><th>Hành động</th></tr></thead>
            <tbody>${list.map(a => `
                <tr class="clickable-row" onclick="event.target.tagName !== 'BUTTON' && navigate('auction-room/${a.id}')">
                    <td><strong>${a.product?.name || `Phiên #${a.id}`}</strong></td>
                    <td class="text-sm">${a.product?.seller?.full_name || 'Hệ thống'}</td>
                    <td style="color:var(--success);font-weight:700">${formatMoney(a.current_price)}</td>
                    <td>${auctionStatusBadge(a)}</td>
                    <td class="text-muted">${formatDate(a.end_time)}</td>
                    <td>
                        <div class="inline-flex gap-2">
                            ${!a.buyer_confirmed && a.status !== 'cancelled' ? `
                                <button class="btn btn-success btn-sm" onclick="confirmReceipt(${a.id})">✅ Đã nhận</button>
                                <button class="btn btn-danger btn-sm" onclick="rejectAuction(${a.id})">❌ Khiếu nại/Hủy</button>
                            ` : (a.status === 'cancelled' ? '<span class="text-danger">✘ Đã hủy</span>' : '<span class="text-success">✔ Đã nhận</span>')}
                        </div>
                    </td>
                </tr>`).join('')}
            </tbody>
        </table></div></div></div>`;
    } catch (err) {
        toast(err.message, 'error');
    }
};

// ========== PROFILE SETTINGS ==========
pages['profile'] = async () => {
    showSidebar();
    $('#app').innerHTML = `<div class="page-header"><h1>⚙️ Cài đặt tài khoản</h1><p>Cập nhật thông tin cá nhân của bạn</p></div>${loading()}`;

    try {
        const res = await api.getMe();
        const u = res.data;
        $('#app').innerHTML = `
        <div class="page-header"><h1>⚙️ Cài đặt tài khoản</h1><p>Cập nhật thông tin cá nhân của bạn</p></div>
        <div class="card" style="max-width: 600px;">
            <div class="card-body">
                <form id="profile-form">
                    <div class="form-group">
                        <label>Họ tên</label>
                        <input type="text" class="form-control" id="prof-name" value="${u.full_name}" required>
                    </div>
                    <div class="form-group">
                        <label>Số điện thoại</label>
                        <input type="text" class="form-control" id="prof-phone" value="${u.phone_number}" required>
                    </div>
                    <div class="form-group">
                        <label>Email (Không thể thay đổi)</label>
                        <input type="email" class="form-control" value="${u.email}" readonly style="background: var(--bg-card); cursor: not-allowed;">
                    </div>
                    <button type="submit" class="btn btn-primary" id="prof-btn">Lưu thay đổi</button>
                </form>
            </div>
        </div>`;

        $('#profile-form').onsubmit = async (e) => {
            e.preventDefault();
            const btn = $('#prof-btn');
            btn.disabled = true;
            try {
                const updated = await api.updateMe({
                    full_name: $('#prof-name').value,
                    phone_number: $('#prof-phone').value
                });
                api.user = updated.data;
                toast('Cập nhật thành công!', 'success');
                navigate('profile');
            } catch (err) {
                toast(err.message, 'error');
            } finally {
                btn.disabled = false;
            }
        };
    } catch (err) {
        toast(err.message, 'error');
    }
};


// ========== ADMIN: USER MANAGEMENT ==========
pages['admin'] = async () => {
    if (!api.isAdmin) return navigate('dashboard');
    showSidebar();
    $('#app').innerHTML = `<div class="page-header"><h1>⚙️ Quản lý Users</h1><p>Tạo, khóa, mở khóa và xóa tài khoản</p></div>${loading()}`;

    try {
        const res = await api.getUsers();
        const list = res.data || [];
        const header = `
        <div class="page-header flex-between">
            <div><h1>⚙️ Quản lý Users</h1><p>${list.length} tài khoản trong hệ thống</p></div>
            <button class="btn btn-primary" onclick="showCreateUserModal()">+ Tạo User</button>
        </div>`;

        $('#app').innerHTML = header + `
        <div class="card"><div class="card-body"><div class="table-wrapper"><table>
            <thead><tr><th>ID</th><th>Họ tên</th><th>Email</th><th>Role</th><th>Trạng thái</th><th>Hành động</th></tr></thead>
            <tbody>${list.map(u => `
                <tr>
                    <td>#${u.id}</td>
                    <td><strong>${u.full_name}</strong></td>
                    <td class="text-muted">${u.email}</td>
                    <td>${roleBadge(u.role)}</td>
                    <td>${u.is_active ? '<span class="badge badge-active">Active</span>' : '<span class="badge badge-locked">Locked</span>'}</td>
                    <td>
                        ${u.role !== 'admin' ? `
                            <div class="inline-flex gap-2">
                                <button class="btn btn-primary btn-sm" onclick='showEditUserModal(${JSON.stringify(u)})'>✏️</button>
                                ${u.is_active
                                    ? `<button class="btn btn-warning btn-sm" onclick="adminLock(${u.id})">🔒 Khóa</button>`
                                    : `<button class="btn btn-success btn-sm" onclick="adminUnlock(${u.id})">🔓 Mở</button>`}
                                <button class="btn btn-danger btn-sm" onclick="adminDelete(${u.id}, '${u.full_name}')">🗑️</button>
                            </div>` : '<span class="text-muted text-sm">—</span>'}
                    </td>
                </tr>`).join('')}
            </tbody>
        </table></div></div></div>`;
    } catch (err) {
        toast(err.message, 'error');
    }
};

function showCreateUserModal() {
    const modal = document.createElement('div');
    modal.className = 'modal-overlay';
    modal.onclick = (e) => { if (e.target === modal) modal.remove(); };
    modal.innerHTML = `
    <div class="modal">
        <h2>+ Tạo User mới</h2>
        <form id="create-user-form">
            <div class="form-group"><label>Họ tên</label><input type="text" class="form-control" id="cu-name" required></div>
            <div class="form-group"><label>Email</label><input type="email" class="form-control" id="cu-email" required></div>
            <div class="form-group"><label>Số điện thoại</label><input type="text" class="form-control" id="cu-phone" required></div>
            <div class="form-group"><label>Mật khẩu</label><input type="password" class="form-control" id="cu-pass" required></div>
            <div class="form-group"><label>Vai trò</label>
                <select class="form-control" id="cu-role">
                    <option value="bidder">Bidder</option>
                    <option value="seller">Seller</option>
                    <option value="admin">Admin</option>
                </select>
            </div>
            <div class="modal-actions">
                <button type="button" class="btn btn-outline" onclick="this.closest('.modal-overlay').remove()">Hủy</button>
                <button type="submit" class="btn btn-primary">Tạo</button>
            </div>
        </form>
    </div>`;
    document.body.appendChild(modal);
    $('#create-user-form').onsubmit = async (e) => {
        e.preventDefault();
        try {
            await api.createUser({ full_name: $('#cu-name').value, email: $('#cu-email').value, phone_number: $('#cu-phone').value, password: $('#cu-pass').value, role: $('#cu-role').value });
            toast('Tạo user thành công!', 'success');
            modal.remove();
            navigate('admin');
        } catch (err) { toast(err.message, 'error'); }
    };
}

function showEditUserModal(user) {
    const modal = document.createElement('div');
    modal.className = 'modal-overlay';
    modal.onclick = (e) => { if (e.target === modal) modal.remove(); };
    modal.innerHTML = `
    <div class="modal">
        <h2>✏️ Sửa User #${user.id}</h2>
        <form id="edit-user-form">
            <div class="form-group"><label>Họ tên</label><input type="text" class="form-control" id="eu-name" value="${user.full_name}" required></div>
            <div class="form-group"><label>Số điện thoại</label><input type="text" class="form-control" id="eu-phone" value="${user.phone_number}" required></div>
            <div class="form-group"><label>Vai trò</label>
                <select class="form-control" id="eu-role">
                    <option value="bidder" ${user.role === 'bidder' ? 'selected' : ''}>Bidder</option>
                    <option value="seller" ${user.role === 'seller' ? 'selected' : ''}>Seller</option>
                    <option value="admin" ${user.role === 'admin' ? 'selected' : ''}>Admin</option>
                </select>
            </div>
            <div class="modal-actions">
                <button type="button" class="btn btn-outline" onclick="this.closest('.modal-overlay').remove()">Hủy</button>
                <button type="submit" class="btn btn-primary">Cập nhật</button>
            </div>
        </form>
    </div>`;
    document.body.appendChild(modal);
    $('#edit-user-form').onsubmit = async (e) => {
        e.preventDefault();
        try {
            await api.updateUser(user.id, { 
                full_name: $('#eu-name').value, 
                phone_number: $('#eu-phone').value, 
                role: $('#eu-role').value 
            });
            toast('Cập nhật user thành công!', 'success');
            modal.remove();
            navigate('admin');
        } catch (err) { toast(err.message, 'error'); }
    };
}

async function adminLock(id) {
    if (!confirm('Khóa tài khoản này?')) return;
    try { await api.lockUser(id); toast('Đã khóa tài khoản', 'success'); navigate('admin'); }
    catch (err) { toast(err.message, 'error'); }
}

async function adminUnlock(id) {
    try { await api.unlockUser(id); toast('Đã mở khóa tài khoản', 'success'); navigate('admin'); }
    catch (err) { toast(err.message, 'error'); }
}

async function adminDelete(id, name) {
    if (!confirm(`Xóa user "${name}"? Hành động này không thể hoàn tác.`)) return;
    try { await api.deleteUser(id); toast('Đã xóa user thành công', 'success'); navigate('admin'); }
    catch (err) { toast(err.message, 'error'); }
}

// ========== CATEGORIES ==========
pages['categories'] = async () => {
    if (api.isLoggedIn) showSidebar(); else showGuestHeader();
    $('#app').innerHTML = `<div class="page-header"><h1>📁 Danh mục sản phẩm</h1><p>Khám phá các loại sản phẩm trong hệ thống</p></div>${loading()}`;

    try {
        const [activeRes, myRes] = await Promise.all([
            api.getCategories('active'),
            api.isLoggedIn ? api.getMyCategories() : Promise.resolve({ data: [] })
        ]);
        
        const activeList = activeRes.data || [];
        const myList = myRes.data || [];
        const isSeller = api.user?.role === 'seller';

        let html = `
        <div class="page-header flex-between">
            <div><h1>📁 Danh mục sản phẩm</h1><p>Có ${activeList.length} danh mục đang hoạt động</p></div>
            ${(isSeller || api.isAdmin) ? `<button class="btn btn-primary" onclick="showCreateCategoryModal()">+ Yêu cầu danh mục mới</button>` : ''}
        </div>
        <div class="auction-grid mb-5">
            ${activeList.map(c => `
                <div class="stat-card clickable-row" onclick="applyCategoryFilterAndNavigate(${c.id})">
                    <div class="stat-label">Danh mục</div>
                    <div class="stat-value text-accent" style="font-size:1.5rem">${c.name}</div>
                    <p class="text-sm text-secondary mt-2">${c.description || 'Không có mô tả'}</p>
                </div>
            `).join('')}
        </div>`;

        if (isSeller && myList.length > 0) {
            html += `
            <div class="card mt-4">
                <div class="card-header"><h2>📝 Yêu cầu của tôi</h2></div>
                <div class="card-body">
                    <div class="table-wrapper"><table>
                        <thead><tr><th>Tên danh mục</th><th>Mô tả</th><th>Trạng thái</th><th>Ghi chú</th></tr></thead>
                        <tbody>${myList.map(c => `
                            <tr>
                                <td><strong>${c.name}</strong></td>
                                <td class="text-muted text-sm">${c.description || '—'}</td>
                                <td>${statusBadge(c.status)}</td>
                                <td class="text-danger text-sm">${c.rejection_reason || '—'}</td>
                            </tr>`).join('')}
                        </tbody>
                    </table></div>
                </div>
            </div>`;
        }

        $('#app').innerHTML = html;
    } catch (err) { toast(err.message, 'error'); }
};

function showCreateCategoryModal() {
    const modal = document.createElement('div');
    modal.className = 'modal-overlay';
    modal.onclick = (e) => { if (e.target === modal) modal.remove(); };
    modal.innerHTML = `
    <div class="modal">
        <h2>📁 Tạo danh mục mới</h2>
        <p class="text-muted text-sm mb-4">${api.isAdmin ? 'Danh mục sẽ tự động được kích hoạt.' : 'Yêu cầu của bạn sẽ được Admin phê duyệt.'}</p>
        <form id="create-cat-form">
            <div class="form-group"><label>Tên danh mục</label><input type="text" class="form-control" id="cat-name" required placeholder="VD: Đồ điện tử"></div>
            <div class="form-group"><label>Mô tả</label><textarea class="form-control" id="cat-desc" rows="3" placeholder="Mô tả ngắn về danh mục..."></textarea></div>
            <div class="modal-actions">
                <button type="button" class="btn btn-outline" onclick="this.closest('.modal-overlay').remove()">Hủy</button>
                <button type="submit" class="btn btn-primary">Gửi yêu cầu</button>
            </div>
        </form>
    </div>`;
    document.body.appendChild(modal);
    $('#create-cat-form').onsubmit = async (e) => {
        e.preventDefault();
        try {
            await api.createCategory({ name: $('#cat-name').value, description: $('#cat-desc').value });
            toast('Gửi yêu cầu thành công!', 'success');
            modal.remove();
            navigate('categories');
        } catch (err) { toast(err.message, 'error'); }
    };
}

// ========== ADMIN: MANAGE CATEGORIES ==========
pages['manage-categories'] = async () => {
    if (!api.isAdmin) return navigate('dashboard');
    showSidebar();
    $('#app').innerHTML = `<div class="page-header"><h1>📂 Phê duyệt danh mục</h1></div>${loading()}`;

    try {
        const res = await api.getCategories(); // Get all
        const list = res.data || [];
        
        $('#app').innerHTML = `
        <div class="page-header"><h1>📂 Phê duyệt danh mục</h1><p>Quản lý yêu cầu tạo danh mục từ Seller</p></div>
        <div class="card"><div class="card-body"><div class="table-wrapper"><table>
            <thead><tr><th>ID</th><th>Tên</th><th>Mô tả</th><th>Trạng thái</th><th>Ghi chú</th><th>Hành động</th></tr></thead>
            <tbody>${list.map(c => `
                <tr>
                    <td>#${c.id}</td>
                    <td><strong>${c.name}</strong></td>
                    <td class="text-muted text-sm">${c.description || '—'}</td>
                    <td>${statusBadge(c.status)}</td>
                    <td class="text-danger text-sm">${c.rejection_reason || '—'}</td>
                    <td>
                        ${c.status === 'pending' ? `
                            <div class="inline-flex gap-2">
                                <button class="btn btn-success btn-sm" onclick="approveCategory(${c.id}, 'active')">✅ Duyệt</button>
                                <button class="btn btn-danger btn-sm" onclick="approveCategory(${c.id}, 'rejected')">❌ Từ chối</button>
                            </div>
                        ` : '—'}
                    </td>
                </tr>`).join('')}
            </tbody>
        </table></div></div></div>`;
    } catch (err) { toast(err.message, 'error'); }
};

window.approveCategory = async (id, status) => {
    let reason = '';
    if (status === 'rejected') {
        reason = prompt('Lý do từ chối:');
        if (reason === null) return;
    }
    try {
        await api.approveCategory(id, status, reason);
        toast('Đã cập nhật trạng thái danh mục', 'success');
        navigate('manage-categories');
    } catch (err) { toast(err.message, 'error'); }
}

window.applyCategoryFilterAndNavigate = (categoryId) => {
    auctionFilters.categories = [categoryId];
    navigate('auctions');
};

// ========== MANAGE PRODUCTS ==========
pages['manage-products'] = async (params = {}) => {
    showSidebar();
    const isSeller = api.user?.role === 'seller';
    const isAdmin = api.user?.role === 'admin';
    const page = parseInt(params.page) || 1;
    const limit = 10;
    
    $('#app').innerHTML = `<div class="page-header"><h1>📦 Quản lý sản phẩm</h1></div>${loading()}`;

    try {
        const queryParams = { page, limit };
        if (isSeller) queryParams.seller_id = api.user.id;
        
        const res = await api.getProducts(queryParams);
        const data = res.data || {};
        const list = data.products || [];
        const total = data.total_count || 0;
        const totalPages = Math.ceil(total / limit);

        $('#app').innerHTML = `
        <div class="page-header flex-between">
            <div>
                <h1>📦 ${isSeller ? 'Sản phẩm của bạn' : 'Tất cả sản phẩm'}</h1>
                <p>${isSeller ? 'Quản lý các sản phẩm bạn đang đấu giá' : 'Danh sách sản phẩm toàn hệ thống'}</p>
            </div>
            ${isSeller ? `<button class="btn btn-primary" onclick="showCreateProductModal()">+ Tạo sản phẩm mới</button>` : ''}
        </div>
        
        <div class="card">
            <div class="card-body">
                ${list.length === 0 ? `
                    <div class="empty-state">
                        <div class="empty-icon">📦</div>
                        <p>${isSeller ? 'Bạn chưa đăng bán sản phẩm nào' : 'Hiện chưa có sản phẩm nào trong hệ thống'}</p>
                    </div>
                ` : `
                    <div class="table-wrapper">
                        <table>
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Ảnh</th>
                                    <th>Tên sản phẩm</th>
                                    <th>Danh mục</th>
                                    ${isAdmin ? '<th>Người bán</th>' : ''}
                                    <th>Trạng thái</th>
                                    <th>Hành động</th>
                                </tr>
                            </thead>
                            <tbody>
                                ${list.map(p => `
                                    <tr>
                                        <td>#${p.id}</td>
                                        <td>
                                            <img src="${p.images?.[0]?.image_url || 'https://via.placeholder.com/50'}" 
                                                 alt="${p.name}" class="table-img" style="width:50px; height:50px; object-fit:cover; border-radius:4px">
                                        </td>
                                        <td><strong>${p.name}</strong></td>
                                        <td>${p.category?.name || '—'}</td>
                                        ${isAdmin ? `<td>${p.seller?.full_name || '—'}</td>` : ''}
                                        <td>${p.auctions?.length > 0 ? auctionStatusBadge(p.auctions.sort((a,b)=>b.id-a.id)[0]) : statusBadge(p.status)}
                                            ${p.rejection_reason ? `<div class="text-xs text-danger mt-1">Lý do: ${p.rejection_reason}</div>` : ''}
                                        </td>
                                        <td>
                                            <div class="inline-flex gap-2">
                                                <button class="btn btn-outline btn-sm" onclick="navigate('auction-detail/${p.id}')">👁️ Xem</button>
                                                
                                                ${/* New: Edit/Delete buttons */ ''}
                                                ${(isSeller || isAdmin) && (!p.auctions || !p.auctions.some(a => ['pending', 'active'].includes(a.status))) ? `
                                                    <button class="btn btn-warning btn-sm" onclick="showEditProductModal(${p.id})">✏️ Sửa</button>
                                                ` : ''}
                                                ${(isSeller || isAdmin) && (!p.auctions || p.auctions.length === 0) ? `
                                                    <button class="btn btn-danger btn-sm" onclick="deleteProduct(${p.id})">🗑️ Xóa</button>
                                                ` : ''}
                                                ${isAdmin ? `
                                                    <button class="btn btn-accent btn-sm" onclick="toggleProductLock(${p.id})">${p.status === 'banned' ? '🔓 Mở khóa' : '🔒 Khóa'}</button>
                                                ` : ''}

                                                ${isSeller && p.status === 'approved' && (!p.auctions || !p.auctions.some(a => ['pending', 'active', 'sold'].includes(a.status) || (a.status === 'ended' && a.winner_id))) ? `
                                                    <button class="btn btn-primary btn-sm" onclick="showCreateAuctionModal(${p.id})">⚖️ Tạo đấu giá</button>
                                                ` : ''}
                                                
                                                ${isSeller && p.auctions?.sort((a,b) => b.id - a.id)[0]?.status === 'ended' && !p.auctions?.sort((a,b) => b.id - a.id)[0]?.seller_confirmed ? `
                                                    <button class="btn btn-success btn-sm" onclick="confirmDelivery(${p.auctions?.sort((a,b) => b.id - a.id)[0].id})">🚚 Xác nhận đã gửi</button>
                                                    <button class="btn btn-danger btn-sm" onclick="rejectAuction(${p.auctions?.sort((a,b) => b.id - a.id)[0].id})">❌ Hủy phiên</button>
                                                ` : ''}

                                                ${isSeller && p.auctions?.sort((a,b) => b.id - a.id)[0]?.status === 'active' ? `
                                                    <button class="btn btn-warning btn-sm" onclick="showExtendAuctionModal(${p.auctions?.sort((a,b) => b.id - a.id)[0].id}, '${p.auctions?.sort((a,b) => b.id - a.id)[0].end_time}')">⏳ Gia hạn</button>
                                                ` : ''}
                                                
                                                ${isAdmin && p.status === 'pending' ? `
                                                    <button class="btn btn-success btn-sm" onclick="approveProduct(${p.id}, 'approved')">✅ Duyệt</button>
                                                    <button class="btn btn-danger btn-sm" onclick="approveProduct(${p.id}, 'rejected')">❌ Từ chối</button>
                                                ` : ''}
                                            </div>
                                        </td>
                                    </tr>
                                `).join('')}
                            </tbody>
                        </table>
                    </div>
                    ${totalPages > 1 ? `
                        <div class="pagination flex-center gap-2 mt-4">
                            <button class="btn btn-sm btn-outline" ${page <= 1 ? 'disabled' : ''} 
                                    onclick="navigate('manage-products', {page: ${page - 1}})">Previous</button>
                            <span class="text-secondary">Trang ${page} / ${totalPages}</span>
                            <button class="btn btn-sm btn-outline" ${page >= totalPages ? 'disabled' : ''} 
                                    onclick="navigate('manage-products', {page: ${page + 1}})">Next</button>
                        </div>
                    ` : ''}
                `}
            </div>
        </div>`;
    } catch (err) { toast(err.message, 'error'); }
};

window.confirmReceipt = async (id) => {
    if (!confirm('Bạn đã nhận được hàng và đồng ý thanh toán cho người bán? Tiền sẽ được chuyển ngay lập tức.')) return;
    try {
        await api.confirmReceipt(id);
        toast('Đã xác nhận nhận hàng! Giao dịch thành công.', 'success');
        navigate('won');
    } catch (err) { toast(err.message, 'error'); }
}

window.rejectAuction = async (id) => {
    const reason = prompt('Bạn chắc chắn muốn hủy/ban phiên đấu giá này? Vui lòng nhập lý do:');
    if (reason === null) return;
    if (!reason.trim()) { toast('Lý do không được để trống', 'warning'); return; }
    try {
        await api.rejectAuction(id, reason);
        toast('Đã hủy phiên đấu giá.', 'info');
        renderPage(window.location.hash.slice(1)); // Refresh
    } catch (err) { toast(err.message, 'error'); }
}

window.confirmDelivery = async (id) => {
    if (!confirm('Bạn xác nhận đã gửi hàng cho người mua?')) return;
    try {
        await api.confirmDelivery(id);
        toast('Đã xác nhận gửi hàng. Đang chờ người mua xác nhận nhận hàng.', 'success');
        navigate('manage-products');
    } catch (err) { toast(err.message, 'error'); }
}

window.showCreateAuctionModal = (productId) => {
    const modal = document.createElement('div');
    modal.className = 'modal-overlay';
    modal.onclick = (e) => { if (e.target === modal) modal.remove(); };
    modal.innerHTML = `
    <div class="modal">
        <h2>⚖️ Tạo phiên đấu giá</h2>
        <form id="create-auction-form">
            <div class="form-group"><label>Giá khởi điểm (₫)</label><input type="number" class="form-control" id="a-start-price" required min="1000"></div>
            <div class="form-group"><label>Bước giá (₫)</label><input type="number" class="form-control" id="a-step-price" value="10000" required min="1000"></div>
            <div class="form-group"><label>Giá mua ngay (₫) - Để trống nếu không có</label><input type="number" class="form-control" id="a-buy-price"></div>
            <div class="form-group"><label>Thời gian bắt đầu</label><input type="datetime-local" class="form-control" id="a-start-time" required></div>
            <div class="form-group"><label>Thời gian kết thúc</label><input type="datetime-local" class="form-control" id="a-end-time" required></div>
            <div class="modal-actions">
                <button type="button" class="btn btn-outline" onclick="this.closest('.modal-overlay').remove()">Hủy</button>
                <button type="submit" class="btn btn-primary" id="a-submit">🚀 Bắt đầu đấu giá</button>
            </div>
        </form>
    </div>`;
    document.body.appendChild(modal);

    const now = new Date();
    const inOneHour = new Date(now.getTime() + 60*60*1000);
    const inOneDay = new Date(now.getTime() + 24*60*60*1000);
    $('#a-start-time').value = inOneHour.toISOString().slice(0, 16);
    $('#a-end-time').value = inOneDay.toISOString().slice(0, 16);

    $('#create-auction-form').onsubmit = async (e) => {
        e.preventDefault();
        try {
            const body = {
                product_id: parseInt(productId),
                start_price: parseFloat($('#a-start-price').value),
                step_price: parseFloat($('#a-step-price').value),
                start_time: new Date($('#a-start-time').value).toISOString(),
                end_time: new Date($('#a-end-time').value).toISOString(),
                status: 'active'
            };
            const buyPrice = parseFloat($('#a-buy-price').value);
            if (buyPrice) body.buy_now_price = buyPrice;

            await api.createAuction(body);
            toast('Đã tạo phiên đấu giá thành công!', 'success');
            modal.remove();
            navigate('auctions');
        } catch (err) { toast(err.message, 'error'); }
    };
}

window.showExtendAuctionModal = (auctionId, currentEndTime) => {
    const modal = document.createElement('div');
    modal.className = 'modal-overlay';
    modal.onclick = (e) => { if (e.target === modal) modal.remove(); };
    
    // Default to +24 hours from current end time
    const current = new Date(currentEndTime);
    const extended = new Date(current.getTime() + 24*60*60*1000);
    
    modal.innerHTML = `
    <div class="modal">
        <h2>⏳ Gia hạn phiên đấu giá #${auctionId}</h2>
        <p class="text-secondary mb-4">Chọn thời gian kết thúc mới cho phiên đấu giá này.</p>
        <form id="extend-auction-form">
            <div class="form-group">
                <label>Thời gian kết thúc mới</label>
                <input type="datetime-local" class="form-control" id="a-new-end-time" required>
            </div>
            <div class="modal-actions">
                <button type="button" class="btn btn-outline" onclick="this.closest('.modal-overlay').remove()">Hủy</button>
                <button type="submit" class="btn btn-warning" id="a-extend-submit">Cập nhật</button>
            </div>
        </form>
    </div>`;
    document.body.appendChild(modal);
    $('#a-new-end-time').value = extended.toISOString().slice(0, 16);

    $('#extend-auction-form').onsubmit = async (e) => {
        e.preventDefault();
        const btn = $('#a-extend-submit');
        btn.disabled = true;
        try {
            const newEndTime = new Date($('#a-new-end-time').value).toISOString();
            await api.extendAuction(auctionId, newEndTime);
            toast('Đã gia hạn phiên đấu giá thành công!', 'success');
            modal.remove();
            navigate('manage-products');
        } catch (err) { toast(err.message, 'error'); }
        finally { btn.disabled = false; }
    };
}

pages['auction-detail'] = async (productId) => {
    // If productId is passed via navigate(path, params), it might be an object
    const id = typeof productId === 'object' ? productId.id : productId;
    if (!id) {
        toast('ID sản phẩm không hợp lệ', 'error');
        return navigate('manage-products');
    }
    showSidebar();
    $('#app').innerHTML = loading();
    try {
        const res = await api.getProduct(id);
        const p = res.data;
        const isAdmin = api.user?.role === 'admin';
        const isSeller = api.user?.role === 'seller' && api.user?.id === p.seller_id;

        $('#app').innerHTML = `
        <div class="page-header flex-between">
            <h1>🔍 Chi tiết sản phẩm #${p.id}</h1>
            <button class="btn btn-outline" onclick="history.back()">⬅ Quay lại</button>
        </div>
        <div class="grid grid-2 gap-4">
            <div class="card">
                <div class="card-body">
                    <div class="product-gallery">
                        ${p.images && p.images.length > 0 ? p.images.map(img => `
                            <img src="${img.image_url}" class="gallery-img mb-2" style="width:100%; border-radius:8px">
                        `).join('') : '<div class="empty-state">Không có ảnh</div>'}
                    </div>
                </div>
            </div>
            <div class="card">
                <div class="card-body">
                    <h2 class="mb-2">${p.name}</h2>
                    <div class="mb-4">${statusBadge(p.status)}</div>
                    <div class="stat-card mb-4" style="background: var(--bg-body)">
                        <div class="stat-label">Danh mục</div>
                        <div class="stat-value text-sm">${p.category?.name || '—'}</div>
                    </div>
                    <div class="stat-card mb-4" style="background: var(--bg-body)">
                        <div class="stat-label">Người bán</div>
                        <div class="stat-value text-sm">${p.seller?.full_name || '—'}</div>
                    </div>
                    <div class="description mb-4">
                        <h3>Mô tả</h3>
                        <p class="text-secondary">${p.description || 'Không có mô tả chi tiết.'}</p>
                    </div>
                    ${p.status === 'approved' && isSeller && (!p.auctions || !p.auctions.some(a => ['pending', 'active', 'sold'].includes(a.status) || (a.status === 'ended' && a.winner_id))) ? `
                        <button class="btn btn-primary btn-block mb-4" onclick="showCreateAuctionModal(${p.id})">⚖️ Tạo phiên đấu giá mới</button>
                    ` : ''}
                    
                    <div class="auction-history mt-4">
                        <h3 class="mb-2">📜 Lịch sử đấu giá</h3>
                        ${!p.auctions || p.auctions.length === 0 ? `
                            <p class="text-secondary text-sm">Chưa có phiên đấu giá nào cho sản phẩm này.</p>
                        ` : `
                            <div class="history-list">
                                ${p.auctions.sort((a,b) => b.id - a.id).map(a => `
                                    <div class="history-item p-3 mb-2 border-radius-sm" style="background: var(--bg-card); border: 1px solid var(--border-color)">
                                         <div class="flex-between mb-1">
                                            <span class="text-sm font-bold">Phiên #${a.id}</span>
                                            ${auctionStatusBadge(a)}
                                        </div>
                                        <div class="flex-between text-xs text-secondary">
                                            <span>Giá cuối: <strong>${formatMoney(a.current_price)}</strong></span>
                                            <span>Ngày: ${new Date(a.end_time).toLocaleDateString()}</span>
                                        </div>
                                    </div>
                                `).join('')}
                            </div>
                        `}
                    </div>
                </div>
            </div>
        </div>
        `;
    } catch (err) { toast(err.message, 'error'); navigate('manage-products'); }
}

window.approveProduct = async (id, status) => {
    let reason = '';
    if (status === 'rejected') {
        reason = prompt('Lý do từ chối:');
        if (reason === null) return;
    }
    try {
        await api.approveProduct(id, status, reason);
        toast('Đã cập nhật trạng thái sản phẩm', 'success');
        navigate('manage-products');
    } catch (err) { toast(err.message, 'error'); }
}

async function showEditProductModal(id) {
    try {
        const [pRes, catRes] = await Promise.all([api.getProduct(id), api.getCategories('active')]);
        const p = pRes.data;
        const categories = catRes.data || [];

        const modal = document.createElement('div');
        modal.className = 'modal-overlay';
        modal.onclick = (e) => { if (e.target === modal) modal.remove(); };

        let imagesHtml = '';
        if (p.images && p.images.length > 0) {
            imagesHtml = `
            <div class="form-group"><label>Ảnh hiện tại</label>
                <div class="image-grid-edit">
                    ${p.images.map(img => `
                        <div class="img-preview-item">
                            <img src="${img.image_url}" style="width: 100px; height: 100px; object-fit: cover; border-radius: 4px;">
                            <button type="button" class="btn-delete-img" onclick="deleteImg(${p.id}, ${img.id}, this)">×</button>
                        </div>
                    `).join('')}
                </div>
            </div>`;
        }

        modal.innerHTML = `
        <div class="modal" style="max-width: 700px;">
            <h2>✏️ Chỉnh sửa sản phẩm #${p.id}</h2>
            <form id="edit-prod-form">
                <div class="form-group"><label>Danh mục</label>
                    <select class="form-control" id="e-p-cat" required>
                        ${categories.map(c => `<option value="${c.id}" ${p.category_id === c.id ? 'selected' : ''}>${c.name}</option>`).join('')}
                    </select>
                </div>
                <div class="form-group"><label>Tên sản phẩm</label><input type="text" class="form-control" id="e-p-name" required value="${p.name}"></div>
                <div class="form-group"><label>Mô tả chi tiết</label><textarea class="form-control" id="e-p-desc" rows="4">${p.description || ''}</textarea></div>
                ${imagesHtml}
                <div class="form-group"><label>Thêm ảnh mới</label><input type="file" class="form-control" id="e-p-images" multiple accept="image/*"></div>
                <div class="modal-actions">
                    <button type="button" class="btn btn-outline" onclick="this.closest('.modal-overlay').remove()">Hủy</button>
                    <button type="submit" class="btn btn-primary" id="e-p-submit">💾 Lưu thay đổi</button>
                </div>
            </form>
        </div>`;
        document.body.appendChild(modal);

        $('#edit-prod-form').onsubmit = async (e) => {
            e.preventDefault();
            const btn = $('#e-p-submit');
            btn.disabled = true;
            try {
                const files = $('#e-p-images').files;
                if (files.length > 0) {
                    const formData = new FormData();
                    formData.append('category_id', $('#e-p-cat').value);
                    formData.append('name', $('#e-p-name').value);
                    formData.append('description', $('#e-p-desc').value);
                    for (let i = 0; i < files.length; i++) {
                        formData.append('images', files[i]);
                    }
                    await api.updateProduct(id, formData);
                } else {
                    await api.updateProduct(id, {
                        category_id: parseInt($('#e-p-cat').value),
                        name: $('#e-p-name').value,
                        description: $('#e-p-desc').value
                    });
                }
                toast('Cập nhật sản phẩm thành công!', 'success');
                modal.remove();
                renderPage(window.location.hash.slice(1));
            } catch (err) { 
                console.error('Error updating product:', err);
                toast('Lỗi cập nhật sản phẩm: ' + err.message, 'error'); 
            }
            finally { btn.disabled = false; }
        };
    } catch (err) { 
        console.error('Error opening edit modal:', err);
        toast('Không thể mở cửa sổ sửa sản phẩm: ' + err.message, 'error'); 
    }
}

window.deleteImg = async (productId, imageId, btn) => {
    if (!confirm('Bạn có muốn xóa ảnh này?')) return;
    try {
        await api.deleteProductImage(productId, imageId);
        toast('Đã xóa ảnh thành công', 'success');
        btn.closest('.img-preview-item').remove();
    } catch (err) { toast(err.message, 'error'); }
};

window.showEditProductModal = showEditProductModal;

window.deleteProduct = async (id) => {
    if (!confirm('Bạn có chắc chắn muốn xóa sản phẩm này? Thao tác này không thể hoàn tác.')) return;
    try {
        await api.deleteProduct(id);
        toast('Đã xóa sản phẩm thành công!', 'success');
        renderPage(window.location.hash.slice(1));
    } catch (err) { toast(err.message, 'error'); }
};

window.toggleProductLock = async (id) => {
    const reason = prompt('Lý do khóa/mở khóa sản phẩm này:');
    if (reason === null) return;
    try {
        await api.lockProduct(id, reason);
        toast('Đã cập nhật trạng thái khóa sản phẩm', 'success');
        renderPage(window.location.hash.slice(1));
    } catch (err) { toast(err.message, 'error'); }
};

// Original showCreateProductModal below
async function showCreateProductModal() {
    try {
        const catRes = await api.getCategories('active');
        const categories = catRes.data || [];

        const modal = document.createElement('div');
        modal.className = 'modal-overlay';
        modal.onclick = (e) => { if (e.target === modal) modal.remove(); };
        modal.innerHTML = `
        <div class="modal" style="max-width: 700px;">
            <h2>📦 Đăng bán sản phẩm mới</h2>
            <form id="create-prod-form">
                <div class="form-group"><label>Danh mục</label>
                    <select class="form-control" id="p-cat" required>
                        <option value="">-- Chọn danh mục --</option>
                        ${categories.map(c => `<option value="${c.id}">${c.name}</option>`).join('')}
                    </select>
                </div>
                <div class="form-group"><label>Tên sản phẩm</label><input type="text" class="form-control" id="p-name" required placeholder="VD: Laptop Dell XPS 15"></div>
                <div class="form-group"><label>Mô tả chi tiết</label><textarea class="form-control" id="p-desc" rows="4"></textarea></div>
                <div class="form-group">
                    <label>Hình ảnh (Có thể chọn nhiều)</label>
                    <input type="file" id="p-images" multiple class="form-control" accept="image/*" style="padding: 10px;">
                    <p class="text-xs text-muted mt-1">Lưu ý: Ảnh sẽ được upload bất đồng bộ lên Cloudinary.</p>
                </div>
                <div class="modal-actions">
                    <button type="button" class="btn btn-outline" onclick="this.closest('.modal-overlay').remove()">Hủy</button>
                    <button type="submit" class="btn btn-primary" id="p-submit">🚀 Tạo sản phẩm</button>
                </div>
            </form>
        </div>`;
        document.body.appendChild(modal);

        $('#create-prod-form').onsubmit = async (e) => {
            e.preventDefault();
            const btn = $('#p-submit');
            btn.disabled = true;
            btn.textContent = 'Đang xử lý...';

            const formData = new FormData();
            formData.append('category_id', $('#p-cat').value);
            formData.append('name', $('#p-name').value);
            formData.append('description', $('#p-desc').value);
            
            const fileInput = $('#p-images');
            for (let i = 0; i < fileInput.files.length; i++) {
                formData.append('images', fileInput.files[i]);
            }

            try {
                await api.createProduct(formData);
                toast('Tạo sản phẩm thành công! Ảnh đang được upload ngầm.', 'success');
                modal.remove();
                navigate('manage-products');
            } catch (err) { toast(err.message, 'error'); }
            finally { btn.disabled = false; btn.textContent = '🚀 Tạo sản phẩm'; }
        };
    } catch (err) { toast('Không thể tải danh mục: ' + err.message, 'error'); }
}

// ========== LOGOUT ==========
$('#btn-logout').onclick = () => {
    api.logout();
    toast('Đã đăng xuất', 'info');
    navigate('login');
};

// ========== INIT ==========
window.addEventListener('hashchange', () => {
    const page = window.location.hash.slice(1) || 'dashboard';
    if (page !== currentPage) navigate(page);
});

// ========== NOTIFICATIONS & WATCHLIST ==========
let notifWs = null;

async function initNotifications() {
    if (!api.isLoggedIn) return;
    
    // Initial count
    try {
        const res = await api.getNotifications(1, 1);
        updateNotifCount(res.data?.total_count || 0);
    } catch (err) { console.error('Failed to fetch notif count', err); }

    const wsUrl = (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/api/v1/ws/notifications';
    notifWs = new WebSocket(wsUrl + '?token=' + api.token);
    
    notifWs.onmessage = (e) => {
        const notif = JSON.parse(e.data);
        toast(`🔔 ${notif.title}: ${notif.content}`, 'info');
        const count = parseInt($('#notif-count').textContent) || 0;
        updateNotifCount(count + 1);
    };

    notifWs.onclose = () => {
        setTimeout(initNotifications, 5000); // Reconnect
    };
}

function updateNotifCount(count) {
    const el = $('#notif-count');
    if (count > 0) {
        el.textContent = count > 99 ? '99+' : count;
        el.classList.remove('hidden');
    } else {
        el.classList.add('hidden');
    }
}

pages['notifications'] = async () => {
    $('#app').innerHTML = `
    <div class="page-header">
        <h1>🔔 Thông báo của bạn</h1>
        <button class="btn btn-outline btn-sm" onclick="window.markAllRead()">Đánh dấu tất cả đã đọc</button>
    </div>
    <div id="notif-content">${loading()}</div>`;

    try {
        const res = await api.getNotifications();
        const notifs = res.data?.notifications || [];
        if (notifs.length === 0) {
            $('#notif-content').innerHTML = emptyState('📭', 'Bạn không có thông báo nào.');
            updateNotifCount(0);
            return;
        }

        $('#notif-content').innerHTML = `
        <div class="notif-list">
            ${notifs.map(n => `
            <div class="notif-item ${n.is_read ? '' : 'unread'}" onclick="window.openNotif(${n.id}, '${n.link}')">
                <div class="notif-icon">${n.type === 'outbid' ? '📉' : '🔔'}</div>
                <div class="notif-content">
                    <div class="notif-title">${n.title}</div>
                    <div class="notif-text">${n.content}</div>
                    <div class="notif-time">${formatDate(n.created_at)}</div>
                </div>
            </div>`).join('')}
        </div>`;
    } catch (err) { toast(err.message, 'error'); }
};

window.openNotif = async (id, link) => {
    try {
        await api.markNotificationAsRead(id);
        if (link && link.includes('auction')) {
            // Extract ID from link like /frontend/auction-detail.html?id=20 or auction-room/20
            const match = link.match(/id=(\d+)/) || link.match(/auction-room\/(\d+)/) || link.match(/auction-detail\/(\d+)/);
            if (match) {
                const auctionId = match[1];
                // Navigate but show join modal immediately for better UX
                window.confirmJoinAuction(auctionId, "Sản phẩm từ thông báo");
                return;
            }
        }
        renderPage('notifications');
    } catch (err) { console.error(err); }
};

window.markAllRead = async () => {
    try {
        await api.markAllNotificationsAsRead();
        toast('Đã đánh dấu tất cả là đã đọc', 'success');
        updateNotifCount(0);
        renderPage('notifications');
    } catch (err) { toast(err.message, 'error'); }
};

pages['watchlist'] = async () => {
    $('#app').innerHTML = `
    <div class="page-header">
        <div>
            <h1>❤️ Danh sách quan tâm</h1>
            <p>Theo dõi và tham gia nhanh các phiên bạn yêu thích</p>
        </div>
    </div>
    <div id="watchlist-content">${loading()}</div>`;

    try {
        const res = await api.getWatchlist();
        const items = res.data || [];
        if (items.length === 0) {
            $('#watchlist-content').innerHTML = emptyState('❤️', 'Bạn chưa quan tâm phiên đấu giá nào.');
            return;
        }

        $('#watchlist-content').innerHTML = `
        <div class="watchlist-grid">
            ${items.map(item => {
                const a = item.auction;
                if (!a) return '';
                return `
                <div class="watchlist-card" onclick="window.confirmJoinAuction(${a.id}, '${a.product?.name || `Phiên #${a.id}`}')">
                    <div class="card-image">
                        <img src="${getProductImage(a.product)}" alt="${a.product?.name}">
                        <div class="card-badges">
                            ${auctionStatusBadge(a)}
                        </div>
                    </div>
                    <div class="card-content">
                        <div class="card-title">${a.product?.name || `Phiên #${a.id}`}</div>
                        <div class="card-info">
                            <span>Người bán:</span>
                            <strong>${a.product?.seller?.full_name || 'Hệ thống'}</strong>
                        </div>
                        <div class="card-info">
                            <span>Kết thúc:</span>
                            <span>${formatDate(a.end_time)}</span>
                        </div>
                        <div class="card-price">
                            ${formatMoney(a.current_price)}
                        </div>
                    </div>
                    <div class="card-footer">
                        <span class="text-xs text-muted">🔨 ${a.bid_count || 0} lượt bid</span>
                        <button class="btn-remove-watch" onclick="event.stopPropagation(); window.toggleWatch(${a.id}, this, true)">
                            Bỏ quan tâm
                        </button>
                    </div>
                </div>`;
            }).join('')}
        </div>`;
    } catch (err) { toast(err.message, 'error'); }
};

window.toggleWatch = async (auctionId, btn, refresh = false) => {
    if (!api.isLoggedIn) {
        toast('Đăng nhập để theo dõi phiên đấu giá', 'warning');
        navigate('login');
        return;
    }
    try {
        const res = await api.toggleWatchlist(auctionId);
        const data = res.data || {};
        toast(data.message, 'success');
        if (refresh) {
            renderPage(currentPage);
        } else {
            btn.classList.toggle('active', data.is_watching);
            btn.textContent = data.is_watching ? '❤️' : '🤍';
        }
    } catch (err) { toast(err.message, 'error'); }
};

// Original Startup below
if (api.isLoggedIn) {
    initNotifications();
}
const initialPage = window.location.hash.slice(1) || (api.isLoggedIn ? 'dashboard' : 'home');
navigate(initialPage);
