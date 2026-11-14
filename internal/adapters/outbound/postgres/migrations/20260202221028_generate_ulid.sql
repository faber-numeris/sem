-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION hex_to_base(hexstr TEXT, base_alphabet TEXT) RETURNS TEXT AS $$
DECLARE
    bytes          BYTEA := ('\x' || hexstr)::BYTEA;
    leading_zeroes INT := 0;
    num            DECIMAL(40,0) := 0;
    base           DECIMAL(40,0) := 1;
    byte_value     INT;
    byte_val       INT;
    byte_values    INT[] DEFAULT ARRAY[]::INT[];
    modulo         INT;
    encoded_str    TEXT := '';
BEGIN
    IF length(hexstr) != 32 THEN
        RAISE EXCEPTION 'Hex-string >%< should be 32 characters long, but is % char(s).', hexstr, length(hexstr);
    END IF;

    FOR hex_index IN REVERSE ((length(hexstr) / 2) - 1)..0 LOOP
        byte_value := get_byte(bytes, hex_index);
        IF byte_value = 0 THEN
            leading_zeroes := leading_zeroes + 1;
        ELSE
            leading_zeroes := 0;
            num := num + (base * byte_value);
        END IF;
        base := base * 256;
    END LOOP;

    WHILE num > 0 LOOP
        modulo := num % length(base_alphabet);
        num := div(num, length(base_alphabet));
        byte_values := array_append(byte_values, modulo);
    END LOOP;

    FOREACH byte_val IN ARRAY byte_values
    LOOP
        encoded_str := SUBSTRING(base_alphabet, byte_val + 1, 1) || encoded_str;
    END LOOP;

    encoded_str := repeat(SUBSTRING(base_alphabet, 1, 1), leading_zeroes) || encoded_str;

    RETURN encoded_str;
END;
$$ LANGUAGE plpgsql IMMUTABLE STRICT;

CREATE OR REPLACE FUNCTION generate_ulid(output_base text default 'base32',
                                         prefix text default NULL,
                                         delimiter text default '_') RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
    timestamp          BYTEA := E'\\000\\000\\000\\000\\000\\000';
    unix_time          BIGINT;
    ulid               BYTEA;
    base32_alphabet    TEXT := '0123456789ABCDEFGHJKMNPQRSTVWXYZ';
    base58_alphabet    TEXT := '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz';
    ulid_base32_length INT := 26;
    return_string      TEXT := '';
BEGIN
    unix_time := (EXTRACT(EPOCH FROM NOW()) * 1000)::BIGINT;
    FOR i IN 0..5 LOOP
        timestamp := SET_BYTE(timestamp, i, (unix_time >> (40 - i * 8))::BIT(8)::INTEGER);
    END LOOP;

    ulid := timestamp || gen_random_bytes(10);
    return_string := SUBSTRING(ulid::text, 3, 32);

    CASE lower(output_base)
        WHEN 'base32' THEN
            return_string := hex_to_base(return_string, base32_alphabet);
            IF ulid_base32_length - length(return_string) > 0 THEN
                return_string := repeat('0', ulid_base32_length - length(return_string)) || return_string;
            END IF;
        WHEN 'base58' THEN
            return_string := hex_to_base(return_string, base58_alphabet);
        WHEN 'uuid' THEN
            -- keep hex string
        ELSE
            RAISE EXCEPTION 'Unsupported output_base %', output_base
            USING HINT = 'Should be base32, base58, or uuid.';
    END CASE;

    IF prefix IS NOT NULL THEN
        return_string := prefix || delimiter || return_string;
    END IF;

    RETURN return_string;
END;
$$;

-- Generates an 18-character Salesforce-compatible ID starting with a 3-char key prefix (e.g. '001', '003', '500').
CREATE OR REPLACE FUNCTION generate_sf_id(key_prefix text default '001') RETURNS varchar(18)
    LANGUAGE plpgsql
    AS $$
DECLARE
    ulid_str TEXT;
    prefix_clean TEXT;
BEGIN
    prefix_clean := rpad(left(coalesce(key_prefix, '001'), 3), 3, '0');
    ulid_str := generate_ulid('base32');
    RETURN (prefix_clean || right(ulid_str, 15))::varchar(18);
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS generate_sf_id(text);
DROP FUNCTION IF EXISTS generate_ulid(text, text, text);
DROP FUNCTION IF EXISTS hex_to_base(text, text);
-- +goose StatementEnd
