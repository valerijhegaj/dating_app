-- для работы версионности scd4
CREATE OR REPLACE FUNCTION dating_data.update_profile() RETURNS TRIGGER AS
$$
DECLARE
    p_id INTEGER;
BEGIN
    p_id := (SELECT profile_id FROM dating_data.profile WHERE user_id = NEW.user_id);
    IF ((SELECT profile_text FROM dating_data.profile WHERE profile_id = p_id) = NEW.profile_text) THEN
        RETURN NULL;
    END IF;

    INSERT INTO dating_data.profile_history (
        profile_id, user_id, profile_text, sex, birthday, name, url, end_time
    )
    SELECT profile_id, user_id, profile_text, sex, birthday, name, url, NOW()
    FROM dating_data.profile
    WHERE profile_id = p_id;

    INSERT INTO dating_data.report_history (
        profile_id, user_id, message
    )
    SELECT profile_id, user_id, message
    FROM dating_data.report
    WHERE profile_id = p_id;

    INSERT INTO dating_data.profile_photo_history (
        profile_id, photo_id
    )
    SELECT profile_id, photo_id
    FROM dating_data.profile_photo
    WHERE profile_id = p_id;

    DELETE FROM dating_data.report
    WHERE profile_id = p_id;

    DELETE FROM dating_data.profile_photo
    WHERE profile_id = p_id;

    DELETE FROM dating_data.profile
    WHERE profile_id = p_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- drop function dating_data.update_profile();

-- работа scd4
CREATE OR REPLACE TRIGGER profile_history_update
BEFORE INSERT ON dating_data.profile
FOR EACH ROW
EXECUTE FUNCTION dating_data.update_profile();

-- drop trigger profile_history_update ON dating_data.profile;

-- Обновить время нахождения в боте
CREATE OR REPLACE FUNCTION dating_data.on_star() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE dating_data.user
    SET last_online = NOW()
    WHERE user_id = NEW.user_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER on_star_trigger
BEFORE INSERT OR UPDATE ON dating_data.starred_users
FOR EACH ROW
EXECUTE FUNCTION dating_data.on_star();
