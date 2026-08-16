import Head from 'next/head';
import Script from 'next/script';
import '../styles/globals.css';

export default function App({ Component, pageProps }) {
  return (
    <>
      <Head>
        <title>META CLASH | Infinite Card Battles</title>
        <meta name="description" content="Generate and battle with any character in the universe." />
        <link rel="icon" href="/favicon.svg" type="image/svg+xml" />
      </Head>
      <Component {...pageProps} />
    </>
  );
}