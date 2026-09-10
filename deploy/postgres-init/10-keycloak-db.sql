-- A Keycloak saját adatbázisa.
--
-- Miért külön adatbázis és nem külön séma: így a mentés és a visszaállítás is
-- szétválasztható. A felhasználói fiókok (Keycloak) és az előfizetések
-- (account szolgáltatás) más ütemben és más okból változnak; egy hibás
-- visszaállítás ne rántsa magával a másikat.
--
-- Ez a fájl CSAK az adatbázis első létrehozásakor fut le (a postgres image
-- docker-entrypoint-initdb.d konvenciója szerint).
SELECT 'CREATE DATABASE keycloak'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'keycloak')\gexec
