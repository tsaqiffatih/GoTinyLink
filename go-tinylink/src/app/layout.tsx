import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const DOMAIN_URL = process.env.NEXT_PUBLIC_DOMAIN_URL;

export const metadata: Metadata = {
  title: "Go-TinyLink: Simplify Your URLs with Our URL Shortener",
  description: "Simplify your links with Go-TinyLink, the easiest-to-use URL shortener. Create short, shareable links in seconds!",
  keywords: [
    "URL shortener",
    "Go-TinyLink",
    "free URL shortener",
    "link shortener",
    "URL shortening tool",
  ],
  authors: [{ name: "Fatih Moh Tsaqif", url: DOMAIN_URL }],
  metadataBase: new URL(process.env.NEXT_PUBLIC_DOMAIN_URL || "http://localhost:3000"),
  openGraph: {
    title: "Go-TinyLink: Simplify Your URLs with Our URL Shortener",
    description:
      "Simplify your links with Go-TinyLink, the easiest-to-use URL shortener. Create short, shareable links in seconds!",
    url: DOMAIN_URL,
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
        alt: "Go-TinyLink: Simplify Your URLs with Our URL Shortener",
      },
    ],
    type: "website",
  },
  robots: {
    index: true,
    follow: true,
  },
  alternates:{
    canonical: DOMAIN_URL,
    languages:{
      en: DOMAIN_URL
    }
  }
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <head>
        {/* Add additional metadata for structured data */}
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "WebSite",
              name: "Go-TinyLink",
              url: DOMAIN_URL,
              description:
                "Simplify your links with Go-TinyLink, the easiest-to-use URL shortener. Create short, shareable links in seconds!",
              potentialAction: {
                "@type": "SearchAction",
                target: `${DOMAIN_URL}/?q={search_term_string}`,
                "query-input": "required name=search_term_string",
              },
            }),
          }}
        />
      </head>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
