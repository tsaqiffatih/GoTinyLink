"use client";

/* eslint-disable @typescript-eslint/no-explicit-any */

import { useState } from "react";
import axios from "axios";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL;

type URLData = {
  url: string;
  shortCode: string;
  accessCount: number;
  expiresAt: string;
};

export default function Home() {
  const [url, setURL] = useState("");
  const [shortenedURL, setShortenedURL] = useState<URLData | null>(null);
  const [error, setError] = useState("");
  const [copySuccess, setCopySuccess] = useState(false);

  const handleShorten = async () => {
    setError("");
    setCopySuccess(false);
    try {
      const { data } = await axios.post(`${API_BASE_URL}/shorten`, { url });
      setShortenedURL(data);
    } catch (err: any) {
      setError(err.response?.data?.error || "An error occurred");
    }
  };

  const handleCopy = () => {
    if (shortenedURL) {
      const link = `${API_BASE_URL}/shorten/${shortenedURL.shortCode}`;
      navigator.clipboard.writeText(link).then(() => {
        setCopySuccess(true);
        setTimeout(() => setCopySuccess(false), 3000);
      });
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-500 via-blue-600 to-indigo-700 flex flex-col items-center justify-center p-6 text-white">
      {/* header section */}
      <header className="text-center mb-8">
        <h1 className="text-4xl font-bold text-white">Go-TinyLink</h1>
        <p className="text-lg font-medium text-gray-100">
          Simplify your links with our easy-to-use URL shortener.
        </p>
      </header>

      {/* main content section */}
      <main className="w-full max-w-md lg:max-w-xl bg-white shadow-lg rounded-lg p-8 text-gray-800">
        <section aria-labelledby="shorten-link-section">
          <h2 id="shorten-link-section" className="text-2xl font-bold text-indigo-700 mb-4">
            Shorten Your Link
          </h2>
          <div className="flex flex-col md:flex-row md:space-x-4">
            <input
              type="url"
              placeholder="Enter a URL to shorten"
              value={url}
              onChange={(e) => setURL(e.target.value)}
              className="input w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 mb-4"
            />
            <button
              onClick={handleShorten}
              className="btn w-full md:w-1/3 bg-indigo-600 text-white px-4 py-3 rounded-lg font-semibold hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              Shorten URL
            </button>
          </div>
          {error && (
            <p className="text-red-500 text-sm mt-4 font-semibold">{error}</p>
          )}
          {shortenedURL && (
            <div className="mt-6 bg-gray-100 p-4 rounded-lg shadow-inner text-left">
              <p className="text-gray-700 font-medium">Your shortened URL:</p>
              <div className="flex items-center gap-2">
                <a
                  href={`${API_BASE_URL}/shorten/${shortenedURL.shortCode}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-indigo-600 font-bold underline break-all"
                >
                  {API_BASE_URL}/shorten/{shortenedURL.shortCode}
                </a>
                <button
                  onClick={handleCopy}
                  className="p-1 text-gray-700 hover:text-indigo-600 tooltip tooltip-right"
                  aria-label="Copy link"
                  data-tip="copy"
                >
                  <svg
                    version="1.1"
                    xmlns="http://www.w3.org/2000/svg"
                    width="25px"
                    height="25px"
                    viewBox="0 0 32 32"
                    className="fill-current"
                  >
                    <path
                      d="M28,4h-2V1c0-0.552-0.448-1-1-1H4C3.448,0,3,0.448,3,1v27c0,0.552,0.448,1,1,1h3v2
                      c0,0.552,0.448,1,1,1h20c0.552,0,1-0.448,1-1V5C29,4.448,28.552,4,28,4z M24,27H5V2h19V27z M27,30H9v-2h15c0.552,0,1-0.448,1-1V6h2
                      V30z M20,9H9V8h11V9z M20,12H9v-1h11V12z M20,15H9v-1h11V15z M20,18H9v-1h11V18z M20,21H9v-1h11V21z"
                    />
                  </svg>
                </button>
              </div>
              {copySuccess && (
                <p className="text-green-500 text-sm mt-2 font-semibold">
                  Link copied!
                </p>
              )}
            </div>
          )}
          <p className="text-sm text-red-600 font-bold underline mt-4">
            Note: All shortened URLs will be automatically deleted after 30 days.
          </p>
        </section>
      </main>

      {/* support section */}
      <aside className="mt-8 bg-white shadow-lg rounded-lg p-6 max-w-md lg:max-w-xl w-full text-center text-gray-800">
        <h2 className="text-2xl font-bold mb-4 text-indigo-700">Support Me ❤️</h2>
        <p className="text-gray-800 mb-4">
          Help me keep Go-TinyLink free and awesome! You can support me by:
        </p>
        <div className="flex flex-col gap-4">
          <a
            href="https://saweria.co/tsaqiffatih"
            target="_blank"
            rel="noopener noreferrer"
            className="bg-yellow-500 text-black px-4 py-2 rounded-lg font-medium hover:bg-yellow-600"
          >
            ☕ Buy me a Coffee
          </a>
          <a
            href="https://github.com/tsaqiffatih/GoTinyLink"
            target="_blank"
            rel="noopener noreferrer"
            className="bg-gray-800 text-white px-4 py-2 rounded-lg font-medium hover:bg-gray-900"
          >
            ⭐ Leave Star on GitHub
          </a>
          <a
            href="https://paypal.me/fatihtsaqif?country.x=ID&locale.x=id_ID"
            target="_blank"
            rel="noopener noreferrer"
            className="bg-green-500 text-black px-4 py-2 rounded-lg font-medium hover:bg-green-600"
          >
            💵 Tip Jar for Fun Projects
          </a>
        </div>
      </aside>

      {/* footer section */}
      <footer className="mt-6 text-sm text-gray-50">
        Powered by <span className="font-bold">GoTinyLink</span>
      </footer>
    </div>
  );
}
