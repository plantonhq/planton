export default function DocsLoading() {
  return (
    <>
      {/* Main content skeleton */}
      <div className="flex-1 min-h-screen min-w-0">
        <div className="px-4 sm:px-6 lg:px-12 py-8 max-w-full animate-pulse">
          <div className="h-8 bg-secondary rounded w-2/3 mb-6" />
          <div className="space-y-3">
            <div className="h-4 bg-secondary/60 rounded w-full" />
            <div className="h-4 bg-secondary/60 rounded w-5/6" />
            <div className="h-4 bg-secondary/60 rounded w-4/6" />
          </div>
          <div className="mt-8 space-y-3">
            <div className="h-6 bg-secondary rounded w-1/3 mb-4" />
            <div className="h-4 bg-secondary/60 rounded w-full" />
            <div className="h-4 bg-secondary/60 rounded w-5/6" />
            <div className="h-4 bg-secondary/60 rounded w-3/6" />
          </div>
          <div className="mt-8 space-y-3">
            <div className="h-6 bg-secondary rounded w-2/5 mb-4" />
            <div className="h-32 bg-secondary/40 rounded-lg border border-border" />
          </div>
        </div>
      </div>

      {/* Right sidebar skeleton */}
      <div className="hidden xl:block sticky top-16 h-[calc(100vh-4rem)] w-80 flex-shrink-0">
        <div className="h-full overflow-y-auto border-l border-border p-4 animate-pulse">
          <div className="h-4 bg-secondary/60 rounded w-2/3 mb-3" />
          <div className="h-3 bg-secondary/40 rounded w-full mb-2" />
          <div className="h-3 bg-secondary/40 rounded w-4/5 mb-2" />
          <div className="h-3 bg-secondary/40 rounded w-3/5 mb-2" />
        </div>
      </div>
    </>
  );
}
