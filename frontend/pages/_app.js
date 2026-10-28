import Head from 'next/head';
import Script from 'next/script';
import '../styles/globals.css';

export default function App({ Component, pageProps }) {
  return (
    <>
      <Head>
        <title>KAIJU CARDS | Infinite Anime Card Battles</title>
        <meta name="description" content="Generate and battle with legendary kaiju, anime, and multiverse characters." />
        <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
      </Head>
      <Component {...pageProps} />
    </>
  );
}
