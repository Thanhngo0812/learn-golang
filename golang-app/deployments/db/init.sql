/* * ============================================================
 * SCRIPT DATABASE V3: HỖ TRỢ DUYỆT CẢ DANH MỤC & SẢN PHẨM
 * ============================================================
 */

-- 1. DỌN DẸP (CLEANUP)
DROP TABLE IF EXISTS bids CASCADE;
DROP TABLE IF EXISTS auctions CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS watchlists CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP TYPE IF EXISTS user_role CASCADE;
DROP TYPE IF EXISTS auction_status CASCADE;
DROP TYPE IF EXISTS product_status CASCADE;
DROP TYPE IF EXISTS category_status CASCADE; -- Mới thêm

-- 2. ĐỊNH NGHĨA ENUM

CREATE TYPE user_role AS ENUM ('admin', 'seller', 'bidder');

-- Trạng thái danh mục (Phục vụ yêu cầu mới của bạn)
CREATE TYPE category_status AS ENUM (
    'pending',    -- Seller tạo, chờ Admin duyệt
    'active',     -- Admin đã duyệt (hoặc Admin tự tạo)
    'rejected'    -- Admin từ chối (Tên danh mục bậy bạ/trùng lặp)
);

CREATE TYPE product_status AS ENUM (
    'draft',      'pending',    'approved',   'rejected',   'sold',    'banned'
);

CREATE TYPE auction_status AS ENUM (
    'pending',    'active',     'ended',      'cancelled',  'banned',     'sold'
);

-- 3. TẠO BẢNG

-- Bảng Users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role user_role DEFAULT 'bidder',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Bảng Categories (Đã nâng cấp)
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Logic duyệt danh mục
    status category_status DEFAULT 'active', -- Nếu Admin tạo thì mặc định active
    created_by INT REFERENCES users(id) ON DELETE SET NULL, -- Ai là người yêu cầu?
    rejection_reason TEXT, -- Lý do từ chối (nếu có)
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Bảng Products
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    seller_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INT REFERENCES categories(id) ON DELETE SET NULL,
    
    name VARCHAR(255) NOT NULL,
    description TEXT,    
    status product_status DEFAULT 'draft',
    rejection_reason TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE product_images (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE, -- Ảnh đại diện
    display_order INT DEFAULT 0 -- Thứ tự hiển thị
);

-- Bảng Auctions
CREATE TABLE auctions (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    
    start_price DECIMAL(15, 2) NOT NULL CHECK (start_price >= 0),
    step_price DECIMAL(15, 2) NOT NULL DEFAULT 10000, 
    current_price DECIMAL(15, 2) NOT NULL DEFAULT 0,
    
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    
    status auction_status DEFAULT 'pending',
    winner_id INT REFERENCES users(id) ON DELETE SET NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    seller_confirmed BOOLEAN DEFAULT FALSE, -- Người bán xác nhận đã giao
    buyer_confirmed BOOLEAN DEFAULT FALSE,  -- Người mua xác nhận đã nhận
    CONSTRAINT check_time_validity CHECK (end_time > start_time)
);
ALTER TABLE auctions 
ADD COLUMN buy_now_price DECIMAL(15, 2), -- Giá mua đứt (NULL = không cho mua đứt)
ADD COLUMN is_auto_extend BOOLEAN DEFAULT TRUE, -- Bật tính năng tự gia hạn
ADD COLUMN extend_time_seconds INT DEFAULT 300, -- Gia hạn 5 phút nếu có bid cuối
ADD COLUMN rejection_reason TEXT; -- Lý do từ chối/ban

-- Bảng Bids
CREATE TABLE bids (
    id SERIAL PRIMARY KEY,
    auction_id INT NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(15, 2) NOT NULL,
    bid_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_positive_amount CHECK (amount > 0)
);

-- Bảng Ví tiền
CREATE TABLE wallets (
    user_id INT PRIMARY KEY REFERENCES users(id),
    balance DECIMAL(15, 2) DEFAULT 0 CHECK (balance >= 0),
    frozen_balance DECIMAL(15, 2) DEFAULT 0, -- Tiền đang bị đóng băng cho các phiên đấu giá
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Bảng Lịch sử giao dịch
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    amount DECIMAL(15, 2) NOT NULL,
    type VARCHAR(50) CHECK (type IN ('deposit', 'withdraw', 'hold', 'refund', 'payment')),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Bảng Thông báo
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'outbid', 'ended', 'status', v.v.
    is_read BOOLEAN DEFAULT FALSE,
    link TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Bảng Theo dõi (Watchlist)
CREATE TABLE watchlists (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    auction_id INT NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, auction_id)
);
-- 4. TẠO INDEX
CREATE INDEX idx_categories_status ON categories(status); -- Admin tìm danh mục chờ duyệt
CREATE INDEX idx_products_seller ON products(seller_id);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_auctions_status ON auctions(status);
CREATE INDEX idx_bids_auction_amount ON bids(auction_id, amount DESC);
CREATE INDEX idx_notifications_user_read ON notifications(user_id, is_read);
CREATE INDEX idx_watchlists_user ON watchlists(user_id);

-- 5. TRIGGER TỰ ĐỘNG CẬP NHẬT GIÁ (Giữ nguyên)
CREATE OR REPLACE FUNCTION check_and_update_bid() RETURNS TRIGGER AS $$
DECLARE
    current_auction RECORD;
    product_seller_id INT;
    bidder_wallet RECORD;
BEGIN
    -- 1. Lấy thông tin và KHÓA dòng dữ liệu (FOR UPDATE) để tránh 2 người bid cùng lúc
    SELECT * INTO current_auction 
    FROM auctions 
    WHERE id = NEW.auction_id 
    FOR UPDATE;

    -- Lấy ID người bán để chặn tự bid (Shill bidding)
    SELECT seller_id INTO product_seller_id 
    FROM products 
    WHERE id = current_auction.product_id;

    -- Lấy thông tin Ví của người đặt giá (khóa dòng để an toàn)
    SELECT * INTO bidder_wallet
    FROM wallets
    WHERE user_id = NEW.user_id
    FOR UPDATE;

    -- 2. Validate cơ bản
    IF current_auction.status != 'active' THEN
        RAISE EXCEPTION 'Phiên đấu giá chưa bắt đầu hoặc đã kết thúc.';
    END IF;

    IF NOW() > current_auction.end_time THEN
        RAISE EXCEPTION 'Phiên đấu giá đã kết thúc.';
    END IF;

    IF NEW.user_id = product_seller_id THEN
        RAISE EXCEPTION 'Bạn không thể tự đấu giá sản phẩm của mình.';
    END IF;

    -- Kiểm tra ví tiền người mua (Tính toán lại nếu cố tình tự đẩy giá)
    IF bidder_wallet IS NULL THEN
        RAISE EXCEPTION 'Lỗi Hệ Thống: Không tìm thấy thông tin ví tiền.';
    END IF;

    IF current_auction.winner_id = NEW.user_id THEN
        -- Nếu là người đang chiến thắng tự đẩy giá lên: Số dư khả dụng được cộng dồn lại phần đã đóng băng của vòng TRƯỚC ĐÓ trong phiên này
        IF (bidder_wallet.balance - bidder_wallet.frozen_balance + current_auction.current_price) < NEW.amount THEN
            RAISE EXCEPTION 'Số dư khả dụng không đủ để tự đẩy giá. Vui lòng nạp thêm tiền.';
        END IF;
    ELSE
        -- Nếu là người chơi bình thường
        IF (bidder_wallet.balance - bidder_wallet.frozen_balance) < NEW.amount THEN
            RAISE EXCEPTION 'Số dư khả dụng không đủ. Vui lòng nạp thêm tiền.';
        END IF;
    END IF;

    -- 3. Logic kiểm tra giá (Xử lý riêng cho người đầu tiên và người đến sau)
    IF current_auction.current_price = 0 THEN
        -- Nếu là người đầu tiên: Chỉ cần lớn hơn hoặc bằng giá khởi điểm
        IF NEW.amount < current_auction.start_price THEN
            RAISE EXCEPTION 'Giá đặt lần đầu tiên phải lớn hơn hoặc bằng giá khởi điểm (% )', current_auction.start_price;
        END IF;
    ELSE
        -- Nếu là người đến sau: Phải lớn hơn giá hiện tại + bước giá
        IF NEW.amount < (current_auction.current_price + current_auction.step_price) THEN
            RAISE EXCEPTION 'Giá đặt tiếp theo phải tối thiểu là %', (current_auction.current_price + current_auction.step_price);
        END IF;

        -- Nếu đã có người Winner trước đó, thì TRẢ LẠI (Unfreeze) tiền cho người thua cuộc
        IF current_auction.winner_id IS NOT NULL THEN
            UPDATE wallets 
            SET frozen_balance = frozen_balance - current_auction.current_price 
            WHERE user_id = current_auction.winner_id;
        END IF;
    END IF;

    -- ĐÓNG BĂNG tiền của người đặt giá mới này (Người đang thắng cuộc)
    UPDATE wallets
    SET frozen_balance = frozen_balance + NEW.amount
    WHERE user_id = NEW.user_id;

    -- 4. Kiểm tra Mua Ngay (Buy Now)
    -- Nếu có giá mua ngay VÀ giá đặt >= giá mua ngay -> Kết thúc luôn
    IF current_auction.buy_now_price IS NOT NULL AND NEW.amount >= current_auction.buy_now_price THEN
        UPDATE auctions 
        SET current_price = NEW.amount,
            winner_id = NEW.user_id,
            status = 'ended',   -- Đổi trạng thái thành Kết thúc (chờ xác nhận)
            end_time = NOW()   -- Kết thúc ngay lập tức
        WHERE id = NEW.auction_id;
        
        RETURN NEW; -- Dừng trigger tại đây
    END IF;

    -- 5. Tự động gia hạn (Anti-Sniping) & Cập nhật giá thường
    -- Logic: Nếu còn < 5 phút và chưa mua ngay -> Cộng giờ
    IF current_auction.is_auto_extend = TRUE AND 
       (EXTRACT(EPOCH FROM (current_auction.end_time - NOW())) < current_auction.extend_time_seconds) THEN
       
       UPDATE auctions 
       SET end_time = end_time + (current_auction.extend_time_seconds || ' seconds')::INTERVAL,
           current_price = NEW.amount,
           winner_id = NEW.user_id
       WHERE id = NEW.auction_id;
    ELSE
       -- Cập nhật giá bình thường
       UPDATE auctions 
       SET current_price = NEW.amount,
           winner_id = NEW.user_id
       WHERE id = NEW.auction_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger gắn vào bảng
DROP TRIGGER IF EXISTS trigger_validate_bid ON bids;
CREATE TRIGGER trigger_validate_bid 
BEFORE INSERT ON bids 
FOR EACH ROW EXECUTE FUNCTION check_and_update_bid();

/* * ============================================================
 * DATA MẪU: DỮ LIỆU PHONG PHÚ ĐỂ TEST TOÀN BỘ HỆ THỐNG
 * ============================================================
 * Tất cả mật khẩu đều là: password123 (đã băm bcrypt)
 */

-- 1. Tạo Users (10 users)
-- Hash bcrypt của "password123": $2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK
INSERT INTO users (full_name, email, phone_number, password_hash, role, is_active) VALUES 
('Admin Hệ Thống', 'admin@sys.com', '0900000001', '$2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK', 'admin', TRUE),
('Nguyễn Văn Bán', 'seller@shop.com', '0900000002', '$2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK', 'seller', TRUE),
('Trần Đại Gia', 'bidderA@example.com', '0900000003', '$2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK', 'bidder', TRUE),
('Lê Phú Ông', 'bidderB@example.com', '0900000004', '$2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK', 'bidder', TRUE),
('Phạm Thị Seller', 'seller2@shop.com', '0900000005', '$2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK', 'seller', TRUE),
('Hoàng Vi Phạm', 'banned@example.com', '0900000006', '$2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK', 'bidder', FALSE),
('Vũ Mới Tạo', 'newbie@example.com', '0900000007', '$2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK', 'bidder', TRUE),
('Trịnh Công Seller', 'seller3@shop.com', '0900000008', '$2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK', 'seller', TRUE),
('Đỗ Mạnh Bidder', 'bidderC@example.com', '0900000009', '$2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK', 'bidder', TRUE),
('Bùi Xuân Bidder', 'bidderD@example.com', '0900000010', '$2a$10$W5FMXC2sGaNMYlrkOYiREevIBu1qBu.KJ7gA0zdG1VUdnS71qy3tK', 'bidder', TRUE);

-- 2. Tạo Ví cho tất cả users (trừ admin)
INSERT INTO wallets (user_id, balance, frozen_balance) VALUES
(2, 100000000, 0),        -- Seller: 100 triệu
(3, 500000000, 0),        -- Đại gia A: 500 triệu
(4, 500000000, 0),        -- Đại gia B: 500 triệu
(5, 50000000, 0),         -- Seller 2: 50 triệu
(6, 30000000, 0),         -- Bị khóa nhưng vẫn có ví
(7, 50000000, 0),         -- Newbie: 50 triệu
(8, 200000000, 0),        -- Seller 3
(9, 500000000, 0),        -- Bidder C: 500 triệu
(10, 100000000, 0);       -- Bidder D: 100 triệu

-- 3. Tạo lịch sử giao dịch mẫu
INSERT INTO transactions (user_id, amount, type, description) VALUES
(3, 500000000, 'deposit', 'Nạp tiền vào ví'),
(4, 500000000, 'deposit', 'Nạp tiền vào ví'),
(9, 500000000, 'deposit', 'Nạp tiền vào ví'),
(10, 100000000, 'deposit', 'Nạp tiền vào ví'),
(3, -5000000, 'withdraw', 'Rút tiền mua sắm'),
(7, 50000000, 'deposit', 'Vốn khởi nghiệp');

-- Tạm thời tắt Trigger để nạp data mẫu (bao gồm các bid cũ, bid đã kết thúc)
ALTER TABLE bids DISABLE TRIGGER trigger_validate_bid;

-- 4. Tạo Categories
INSERT INTO categories (name, description, status, created_by) VALUES 
('Điện tử', 'Laptop, điện thoại, máy tính bảng', 'active', 1),
('Đồ Gia Dụng', 'Nồi niêu xoong chảo', 'active', 1),
('Thời Trang', 'Quần áo, giày dép, phụ kiện', 'active', 1),
('Đồ Cổ', 'Tranh, tượng, đồ sưu tầm', 'active', 1),
('Xe cộ', 'Ô tô, xe máy cũ', 'active', 1),
('Sách & Nghệ thuật', 'Sách hiếm, tác phẩm nghệ thuật', 'active', 1);

-- 5. Tạo sản phẩm phong phú (Hơn 15 sản phẩm)
INSERT INTO products (seller_id, category_id, name, description, status) VALUES 
(2, 1, 'Laptop Dell XPS 13', 'i7 Gen 12, RAM 16GB, SSD 512GB, Màn 4K', 'approved'),
(2, 1, 'iPhone 15 Pro Max', 'Titan Tự Nhiên, 256GB, VN/A, Fullbox', 'approved'),
(2, 2, 'Nồi cơm điện Panasonic', 'Cao tần IH, 1.8L, mới 95%', 'approved'),
(5, 3, 'Áo Khoác Burberry', 'Size L, họa tiết vintage, hàng auth', 'approved'),
(5, 4, 'Tranh Sơn Dầu 1960', 'Tác phẩm phong cảnh Bắc Bộ', 'approved'),
(8, 5, 'Honda Cub 50cc 1980', 'Xe zin, máy êm, giấy tờ đầy đủ', 'approved'),
(8, 5, 'Vespa Sprint 2022', 'Mới đi 2000km, màu xám xi măng', 'approved'),
(2, 1, 'MacBook Pro M3 Max', '14 inch, RAM 36GB, SSD 1TB, Apple Care+', 'approved'),
(5, 3, 'Đồng hồ Omega Seamaster', 'Máy automatic, demi vàng 18k', 'approved'),
(8, 6, 'Cuốn Sách Cổ 1920', 'Ngữ pháp tiếng Việt xưa, hiếm', 'approved'),
(2, 1, 'Bàn phím cơ Custom', 'Keychron Q1, Switch Holy Panda, Keycap PBT', 'approved'),
(5, 4, 'Bình Gốm Chu Đậu', 'Hoa văn vẽ tay, hàng sưu tầm', 'approved'),
(2, 2, 'Máy lọc không khí Dyson', 'Model 2023, lọc bụi mịn và formaldehyde', 'approved'),
(8, 1, 'Máy ảnh Sony A7IV', 'Fullframe, 33MP, mới chụp 1000 shot', 'approved'),
(5, 3, 'Túi xách Chanel Classic', 'Size Medium, Caviar leather, kẹp chì 2022', 'approved');

-- 6. Tạo các phiên đấu giá đa dạng (Hơn 12 phiên)
INSERT INTO auctions (product_id, start_price, step_price, current_price, start_time, end_time, status, buy_now_price) VALUES 
-- Đang hoạt động (active)
(1, 15000000, 500000, 15000000, NOW() - INTERVAL '1 day', NOW() + INTERVAL '2 days', 'active', 25000000),
(2, 22000000, 1000000, 22000000, NOW() - INTERVAL '6 hours', NOW() + INTERVAL '5 days', 'active', 35000000),
(3, 800000, 50000, 800000, NOW() - INTERVAL '12 hours', NOW() + INTERVAL '3 days', 'active', 2000000),
(4, 5000000, 200000, 5000000, NOW(), NOW() + INTERVAL '4 days', 'active', 10000000),
(5, 30000000, 2000000, 30000000, NOW() - INTERVAL '1 day', NOW() + INTERVAL '1 day', 'active', 60000000),
(6, 12000000, 500000, 12000000, NOW(), NOW() + INTERVAL '6 days', 'active', 20000000),
(7, 45000000, 1000000, 45000000, NOW() - INTERVAL '2 days', NOW() + INTERVAL '3 days', 'active', 65000000),
(8, 55000000, 2000000, 55000000, NOW(), NOW() + INTERVAL '7 days', 'active', 80000000),
(9, 70000000, 3000000, 70000000, NOW() - INTERVAL '1 day', NOW() + INTERVAL '2 days', 'active', 120000000),
(10, 1500000, 100000, 1500000, NOW(), NOW() + INTERVAL '5 days', 'active', 5000000),
(11, 2000000, 100000, 2000000, NOW() - INTERVAL '4 hours', NOW() + INTERVAL '2 days', 'active', 4500000),
(14, 40000000, 1000000, 40000000, NOW(), NOW() + INTERVAL '10 days', 'active', 60000000);

-- 7. Tạo một số Bid ban đầu
INSERT INTO bids (auction_id, user_id, amount) VALUES
(1, 3, 15500000), (1, 4, 16000000),
(2, 9, 23000000), (2, 3, 24000000),
(3, 10, 850000), (3, 7, 900000), (3, 10, 950000),
(5, 4, 32000000), (5, 9, 34000000),
(7, 9, 46000000), (7, 3, 47000000), (7, 4, 48000000),
(9, 4, 73000000), (9, 9, 76000000),
(11, 10, 2100000), (11, 7, 2200000);

-- 8. Tạo dữ liệu đã thắng (Auction ended)
INSERT INTO products (seller_id, category_id, name, description, status) 
VALUES (2, 1, 'iPad Pro M2', '11 inch, 128GB, Wifi, Xám Không Gian', 'approved');
-- Insert auction đã ended
-- ID product 16
INSERT INTO auctions (product_id, start_price, step_price, current_price, start_time, end_time, status, winner_id) 
VALUES (16, 12000000, 500000, 18000000, NOW() - INTERVAL '5 days', NOW() - INTERVAL '1 day', 'ended', 3);
-- Thêm bid thắng (auction_id 13 - giả sử)
INSERT INTO bids (auction_id, user_id, amount, bid_time) 
VALUES (13, 4, 17500000, NOW() - INTERVAL '2 days'), (13, 3, 18000000, NOW() - INTERVAL '1 day');

-- 9. MOCK DỮ LIỆU ĐỂ TEST TỐI ƯU HÓA TRUY VẤN SẢN PHẨM HOT
DO $$
DECLARE
    i INT;
    new_product_id INT;
    new_auction_id INT;
BEGIN
    FOR i IN 1..10 LOOP
        INSERT INTO products (seller_id, category_id, name, description, status)
        VALUES (2, 1, 'Sản Phẩm Hot Thứ ' || i, 'Mô tả chi tiết sản phẩm hot thứ ' || i || '. Sản phẩm chất lượng cao, được nhiều người quan tâm.', 'approved')
        RETURNING id INTO new_product_id;

        INSERT INTO auctions (product_id, start_price, step_price, current_price, start_time, end_time, status)
        VALUES (new_product_id, 1000000, 50000, 1000000, NOW(), NOW() + INTERVAL '7 days', 'active')
        RETURNING id INTO new_auction_id;

        -- Tạo Bids: SP1 = 15 Bid, SP2 = 14 Bid, ...
        FOR j IN 1..(16 - i) LOOP
            INSERT INTO bids (auction_id, user_id, amount) 
            VALUES (new_auction_id, CASE WHEN j % 2 = 0 THEN 3 ELSE 4 END, 1000000 + (j * 50000));
        END LOOP;
    END LOOP;
END $$;

-- Thêm 2 phiên đã kết thúc để test luồng xác nhận (Sửa từ USER)
DO $$
DECLARE
    p_id_1 INT;
    p_id_2 INT;
BEGIN
    INSERT INTO products (seller_id, category_id, name, description, status)
    VALUES (2, 1, 'Máy Ảnh Canon EOS (Test Won)', 'Sản phẩm test cho luồng thắng và xác nhận.', 'approved')
    RETURNING id INTO p_id_1;

    INSERT INTO auctions (product_id, start_price, step_price, current_price, start_time, end_time, status, winner_id, seller_confirmed, buyer_confirmed)
    VALUES (p_id_1, 2000000, 100000, 2500000, NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 hour', 'ended', 3, false, false);

    INSERT INTO products (seller_id, category_id, name, description, status)
    VALUES (2, 1, 'Đồng Hồ Rolex (Test Won)', 'Sản phẩm test cho luồng thắng và xác nhận.', 'approved')
    RETURNING id INTO p_id_2;

    INSERT INTO auctions (product_id, start_price, step_price, current_price, start_time, end_time, status, winner_id, seller_confirmed, buyer_confirmed)
    VALUES (p_id_2, 5000000, 500000, 6000000, NOW() - INTERVAL '3 days', NOW() - INTERVAL '2 hours', 'ended', 3, false, false);

    -- TỰ ĐỘNG TÍNH TOÁN LẠI FROZEN BALANCE CHO TẤT CẢ USER
    -- Logic: Frozen = Tổng current_price của các phiên (Active + Ended chưa xong confirm) mà user đang dẫn đầu/thắng
    UPDATE wallets w
    SET frozen_balance = (
        SELECT COALESCE(SUM(current_price), 0)
        FROM auctions a
        WHERE a.winner_id = w.user_id
        AND (
            a.status = 'active' 
            OR (a.status = 'ended' AND (a.seller_confirmed = false OR a.buyer_confirmed = false))
        )
    );
END $$;

-- Bạt lại Trigger sau khi nạp xong data mẫu
ALTER TABLE bids ENABLE TRIGGER trigger_validate_bid;