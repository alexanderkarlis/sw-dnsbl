SET @ip_address = "127.0.0.255", @response_code = "NXDOMAIN", @updated_at = "111111111";
INSERT INTO ip_details(
        ip_address,
        response_code,
        updated_at
    )
VALUES
    (@ip_address, @response_code, @updated_at)
ON DUPLICATE KEY UPDATE
    response_code = @response_code,
    updated_at = @updated_at;