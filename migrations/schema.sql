CREATE TABLE chat_subscriptions
(
    subscription_id SERIAL PRIMARY KEY,
    chat_id         BIGINT,
    user_id         BIGINT,
    area            VARCHAR(32),
    keywords        VARCHAR(256),
    experience      VARCHAR(128),
    created_at      TIMESTAMP,
    CONSTRAINT unique_subscription UNIQUE (chat_id, area, keywords, experience)
);

CREATE TABLE chat_sent_vacancies
(
    sent_id         SERIAL PRIMARY KEY,
    subscription_id INT REFERENCES chat_subscriptions (subscription_id) ON DELETE CASCADE,
    vacancy_id      VARCHAR(128),
    created_at      TIMESTAMP
);

CREATE TABLE chat_trees
(
    chat_tree_id    SERIAL PRIMARY KEY,
    chat_id         BIGINT,
    serialized_tree JSONB,
    created_at      TIMESTAMP
);

SELECT DISTINCT subscriptions_ids,
                user_ids,
                chat_ids,
                area,
                keywords,
                experience
FROM (SELECT subscription_id,
             user_id,
             chat_id,
             area,
             experience,
             TRIM(REGEXP_REPLACE(LOWER(keywords), '\s+', ' ', 'g'))               as keywords,
             ARRAY_AGG(subscription_id) OVER (PARTITION BY area, LOWER(keywords)) as subscriptions_ids,
             ARRAY_AGG(user_id) OVER (PARTITION BY area, LOWER(keywords))         as user_ids,
             ARRAY_AGG(chat_id) OVER (PARTITION BY area, LOWER(keywords))         as chat_ids
      FROM chat_subscriptions) as SUBQUERY;