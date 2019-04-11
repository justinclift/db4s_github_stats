--
-- PostgreSQL database dump
--

-- Dumped from database version 10.7
-- Dumped by pg_dump version 10.7

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: github_download_counts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.github_download_counts (
    count_id integer NOT NULL,
    asset integer,
    download_count integer NOT NULL,
    count_timestamp integer NOT NULL
);


--
-- Name: github_download_counts_count_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.github_download_counts_count_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: github_download_counts_count_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.github_download_counts_count_id_seq OWNED BY public.github_download_counts.count_id;


--
-- Name: github_download_timestamps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.github_download_timestamps (
    timestamp_id integer NOT NULL,
    count_timestamp timestamp with time zone NOT NULL
);


--
-- Name: github_download_timestamps_timestamp_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.github_download_timestamps_timestamp_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: github_download_timestamps_timestamp_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.github_download_timestamps_timestamp_id_seq OWNED BY public.github_download_timestamps.timestamp_id;


--
-- Name: github_release_assets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.github_release_assets (
    asset_id integer NOT NULL,
    asset_name text NOT NULL
);


--
-- Name: github_release_assets_asset_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.github_release_assets_asset_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: github_release_assets_asset_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.github_release_assets_asset_id_seq OWNED BY public.github_release_assets.asset_id;


--
-- Name: github_download_counts count_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_download_counts ALTER COLUMN count_id SET DEFAULT nextval('public.github_download_counts_count_id_seq'::regclass);


--
-- Name: github_download_timestamps timestamp_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_download_timestamps ALTER COLUMN timestamp_id SET DEFAULT nextval('public.github_download_timestamps_timestamp_id_seq'::regclass);


--
-- Name: github_release_assets asset_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_release_assets ALTER COLUMN asset_id SET DEFAULT nextval('public.github_release_assets_asset_id_seq'::regclass);


--
-- Name: github_download_timestamps github_download_timestamps_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_download_timestamps
    ADD CONSTRAINT github_download_timestamps_pk PRIMARY KEY (timestamp_id);


--
-- Name: github_release_assets github_release_assets_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_release_assets
    ADD CONSTRAINT github_release_assets_pk PRIMARY KEY (asset_id);


--
-- Name: github_download_counts_count_id_uindex; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX github_download_counts_count_id_uindex ON public.github_download_counts USING btree (count_id);


--
-- Name: github_download_timestamps_download_timestamp_uindex; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX github_download_timestamps_download_timestamp_uindex ON public.github_download_timestamps USING btree (count_timestamp);


--
-- Name: github_download_timestamps_timestamp_id_uindex; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX github_download_timestamps_timestamp_id_uindex ON public.github_download_timestamps USING btree (timestamp_id);


--
-- Name: github_release_assets_asset_id_uindex; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX github_release_assets_asset_id_uindex ON public.github_release_assets USING btree (asset_id);


--
-- Name: github_release_assets_asset_name_uindex; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX github_release_assets_asset_name_uindex ON public.github_release_assets USING btree (asset_name);


--
-- Name: github_download_counts github_download_counts_github_download_timestamps_timestamp_id_; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_download_counts
    ADD CONSTRAINT github_download_counts_github_download_timestamps_timestamp_id_ FOREIGN KEY (count_timestamp) REFERENCES public.github_download_timestamps(timestamp_id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: github_download_counts github_download_counts_github_release_assets_asset_id_fk; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.github_download_counts
    ADD CONSTRAINT github_download_counts_github_release_assets_asset_id_fk FOREIGN KEY (asset) REFERENCES public.github_release_assets(asset_id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- PostgreSQL database dump complete
--

