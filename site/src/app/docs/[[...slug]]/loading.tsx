export default function DocsLoading() {
  return (
    <>
      {/* Main Content Area - skeleton */}
      <div className="flex-1 min-h-screen overflow-x-hidden">
        <div className="px-4 sm:px-6 lg:px-12 py-8 max-w-full animate-pulse">
          {/* Title skeleton */}
          <div className="h-8 bg-slate-800/50 rounded w-2/3 mb-6" />
          {/* Paragraph skeletons */}
          <div className="space-y-3">
            <div className="h-4 bg-slate-800/30 rounded w-full" />
            <div className="h-4 bg-slate-800/30 rounded w-5/6" />
            <div className="h-4 bg-slate-800/30 rounded w-4/6" />
          </div>
          <div className="mt-8 space-y-3">
            <div className="h-6 bg-slate-800/50 rounded w-1/3 mb-4" />
            <div className="h-4 bg-slate-800/30 rounded w-full" />
            <div className="h-4 bg-slate-800/30 rounded w-5/6" />
            <div className="h-4 bg-slate-800/30 rounded w-3/6" />
          </div>
          <div className="mt-8 space-y-3">
            <div className="h-6 bg-slate-800/50 rounded w-2/5 mb-4" />
            <div className="h-32 bg-slate-800/20 rounded-lg border border-purple-900/20" />
          </div>
        </div>
      </div>

      {/* Right Sidebar - skeleton */}
      <div className="hidden xl:block sticky top-16 h-[calc(100vh-4rem)] w-80 flex-shrink-0">
        <div className="h-full overflow-y-auto bg-slate-950 border-l border-purple-900/30 p-4 animate-pulse">
          <div className="h-4 bg-slate-800/30 rounded w-2/3 mb-3" />
          <div className="h-3 bg-slate-800/20 rounded w-full mb-2" />
          <div className="h-3 bg-slate-800/20 rounded w-4/5 mb-2" />
          <div className="h-3 bg-slate-800/20 rounded w-3/5 mb-2" />
        </div>
      </div>
    </>
  );
}
