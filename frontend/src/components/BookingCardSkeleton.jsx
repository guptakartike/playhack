import React from 'react';

export default function BookingCardSkeleton() {
  return (
    <div className="rounded-2xl bg-[#14141E] border border-[#2A2A3C] p-5 sm:p-6 space-y-4 animate-pulse">
      <div className="flex items-start justify-between">
        <div className="space-y-2">
          <div className="h-3 w-24 bg-[#2A2A3C] rounded" />
          <div className="h-5 w-48 bg-[#2A2A3C] rounded" />
        </div>
        <div className="h-6 w-20 bg-[#2A2A3C] rounded-full" />
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2 border-t border-[#1F1F2E]">
        <div className="h-4 w-36 bg-[#2A2A3C] rounded" />
        <div className="h-4 w-24 bg-[#2A2A3C] rounded" />
      </div>

      <div className="pt-2 flex justify-end">
        <div className="h-8 w-28 bg-[#2A2A3C] rounded-xl" />
      </div>
    </div>
  );
}
