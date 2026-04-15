@if (auth()->user()?->role === \App\Enum\User\RoleUser::STAFF)
    @php
        $inviteCode = auth()->user()?->affiliateProfile?->ref_code;
    @endphp

    <div class="hidden items-center gap-2 md:flex">
        <div class="inline-flex items-center gap-2 rounded-full border border-rose-200 bg-rose-50 px-3 py-1.5 text-xs font-semibold text-rose-700 shadow-sm">
            <span class="material-symbols-outlined text-[14px]">badge</span>
            <span>Mã mời</span>
            <span class="rounded-full bg-white px-2 py-0.5 font-black tracking-wide text-slate-900">
                {{ filled($inviteCode) ? $inviteCode : 'Chưa tạo' }}
            </span>
            @if (filled($inviteCode))
                <button
                    type="button"
                    class="rounded-full bg-white px-2 py-0.5 font-bold text-rose-600 transition hover:bg-rose-100"
                    x-data
                    x-on:click="navigator.clipboard.writeText(@js($inviteCode))"
                >
                    Copy
                </button>
            @endif
        </div>
    </div>
@endif
