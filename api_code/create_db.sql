SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;
SET default_tablespace = '';
SET default_table_access_method = heap;
CREATE TABLE public.account (
    "userID" integer NOT NULL,
    "firstName" character varying(255) NOT NULL,
    "lastName" character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    premium boolean NOT NULL,
    business boolean NOT NULL
);
CREATE TABLE public.listens (
    "songID" integer NOT NULL,
    "userID" integer NOT NULL
);
CREATE TABLE public.loc (
    "userID" integer NOT NULL,
    latitude numeric(8,6) NOT NULL,
    longitude numeric(9,6) NOT NULL
);
CREATE TABLE public.music (
    "songID" integer NOT NULL,
    "songName" character varying NOT NULL,
    "artist" character varying NOT NULL,
    "length" time without time zone NOT NULL
);
ALTER TABLE ONLY public.music
    ADD CONSTRAINT pk_music PRIMARY KEY ("songID");
ALTER TABLE ONLY public.account
    ADD CONSTRAINT pk_user PRIMARY KEY ("userID");
ALTER TABLE ONLY public.listens
    ADD CONSTRAINT "fk_listen_songID" FOREIGN KEY ("songID") REFERENCES public.music("songID");
ALTER TABLE ONLY public.listens
    ADD CONSTRAINT "fk_listen_userID" FOREIGN KEY ("userID") REFERENCES public.account("userID");
ALTER TABLE ONLY public.loc
    ADD CONSTRAINT "fk_location_userID" FOREIGN KEY ("userID") REFERENCES public.account("userID");