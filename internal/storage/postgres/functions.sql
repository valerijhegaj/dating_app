-- добавить загруженное фото к анкете
CREATE OR REPLACE PROCEDURE dating_data.add_photo(profile_id_in INTEGER, photo_id_in INTEGER) AS
$$
BEGIN
    INSERT INTO dating_data.profile_photo
    (profile_id, photo_id) VALUES (profile_id_in, photo_id_in);
END;
$$ language plpgsql;

-- добавить новое фото к анкете
CREATE OR REPLACE PROCEDURE dating_data.add_new_photo(profile_id_in INTEGER, photo_URL TEXT) AS
$$
DECLARE
    p_id INTEGER;
BEGIN
    p_id := (SELECT photo_id FROM dating_data.photo WHERE image_url = photo_URL);
    IF (p_id IS NULL) THEN
        INSERT INTO dating_data.photo (image_url) VALUES (photo_URL);
        p_id := (SELECT photo_id FROM dating_data.photo WHERE image_url = photo_URL);
    END IF;
    CALL dating_data.add_photo(profile_id_in, p_id);
END;
$$ language plpgsql;

-- создать анкету с загруженным фото или без фото
CREATE OR REPLACE FUNCTION dating_data.create_profile(user_id_in INTEGER, profile_text_in TEXT, sex_in BOOLEAN,
                                                      birthday_in TIMESTAMP, name_in VARCHAR(255),
                                                      photo_id_in INTEGER DEFAULT 0) RETURNS INTEGER AS
$$
DECLARE
    p_id INTEGER;
BEGIN
    INSERT INTO dating_data.profile
    (user_id, profile_text, sex, birthday, name) VALUES (user_id_in, profile_text_in, sex_in, birthday_in, name_in);

    p_id := (SELECT profile_id FROM dating_data.profile WHERE user_id = user_id_in);

    IF (photo_id_in != 0) THEN
        CALL dating_data.add_photo(p_id, photo_id_in);
    END IF;
    RETURN p_id;
END;
$$ language plpgsql;

-- создать анкету с новым фото
CREATE OR REPLACE FUNCTION dating_data.create_profile_with_new_photo(user_id_in INTEGER, profile_text_in TEXT,
                                                                     sex_in BOOLEAN, birthday_in TIMESTAMP,
                                                                     name_in VARCHAR(255), photo_URL TEXT)
    RETURNS INTEGER AS
$$
DECLARE
    p_id INTEGER;
BEGIN
    INSERT INTO dating_data.profile
    (user_id, profile_text, sex, birthday, name) VALUES (user_id_in, profile_text_in, sex_in, birthday_in, name_in);

    p_id := (SELECT profile_id FROM dating_data.profile WHERE user_id = user_id_in);

    CALL dating_data.add_new_photo(p_id, photo_URL);
    RETURN p_id;
END;
$$ language plpgsql;

-- получить подобранную пору
CREATE OR REPLACE FUNCTION dating_data.get_indexed(userID INTEGER) RETURNS INTEGER AS
$$
DECLARE
    i_id INTEGER;
BEGIN
    i_id := (SELECT indexed_user_id FROM dating_data.indexed_users WHERE user_id = userID LIMIT 1);
    DELETE FROM dating_data.indexed_users WHERE user_id = userID AND indexed_user_id = i_id;
    INSERT INTO dating_data.starred_users
    (user_id, starred_user_id, viewed, is_liked, time) VALUES
    (userID, i_id, false, false, NOW());

    RETURN i_id;
END;
$$ language plpgsql;

