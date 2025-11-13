"use client";

import DroneMap from "@/components/DroneMap";

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-start bg-zinc-50 font-sans dark:bg-black">
      <main className="flex flex-col w-full max-w-5xl items-center py-8 px-4">
        <h1 className="text-3xl font-semibold text-center text-black dark:text-zinc-50 mb-4">
          Altivion — Live Drone Map
        </h1>

        <p className="text-zinc-600 dark:text-zinc-400 text-center mb-6 max-w-2xl">
          Real-time drone tracking powered by TimescaleDB and FastAPI.
        </p>

        <div className="w-full rounded-xl overflow-hidden shadow-lg border border-zinc-200 dark:border-zinc-800">
          <DroneMap />
        </div>

        <footer className="mt-6 text-xs text-zinc-500 dark:text-zinc-400 text-center">
          Data updates automatically every few seconds.
        </footer>
      </main>
    </div>
  );
}
