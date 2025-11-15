-- +goose Up
-- +goose StatementBegin
INSERT INTO devices (id, name, brand, state, created_at) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'MacBook Pro 16"', 'Apple', 0, CURRENT_TIMESTAMP - INTERVAL '30 days'),
    ('550e8400-e29b-41d4-a716-446655440002', 'ThinkPad X1 Carbon', 'Lenovo', 1, CURRENT_TIMESTAMP - INTERVAL '25 days'),
    ('550e8400-e29b-41d4-a716-446655440003', 'Dell XPS 15', 'Dell', 0, CURRENT_TIMESTAMP - INTERVAL '20 days'),
    ('550e8400-e29b-41d4-a716-446655440004', 'Surface Laptop 5', 'Microsoft', 1, CURRENT_TIMESTAMP - INTERVAL '15 days'),
    ('550e8400-e29b-41d4-a716-446655440005', 'iPhone 15 Pro', 'Apple', 0, CURRENT_TIMESTAMP - INTERVAL '10 days'),
    ('550e8400-e29b-41d4-a716-446655440006', 'Galaxy S24 Ultra', 'Samsung', 2, CURRENT_TIMESTAMP - INTERVAL '45 days'),
    ('550e8400-e29b-41d4-a716-446655440007', 'iPad Pro 12.9"', 'Apple', 1, CURRENT_TIMESTAMP - INTERVAL '5 days'),
    ('550e8400-e29b-41d4-a716-446655440008', 'Pixel 8 Pro', 'Google', 0, CURRENT_TIMESTAMP - INTERVAL '8 days'),
    ('550e8400-e29b-41d4-a716-446655440009', 'ROG Strix G15', 'Asus', 0, CURRENT_TIMESTAMP - INTERVAL '12 days'),
    ('550e8400-e29b-41d4-a716-446655440010', 'Alienware m17', 'Dell', 2, CURRENT_TIMESTAMP - INTERVAL '60 days')
ON CONFLICT (id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM devices WHERE id IN (
    '550e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440003',
    '550e8400-e29b-41d4-a716-446655440004',
    '550e8400-e29b-41d4-a716-446655440005',
    '550e8400-e29b-41d4-a716-446655440006',
    '550e8400-e29b-41d4-a716-446655440007',
    '550e8400-e29b-41d4-a716-446655440008',
    '550e8400-e29b-41d4-a716-446655440009',
    '550e8400-e29b-41d4-a716-446655440010'
);
-- +goose StatementEnd
