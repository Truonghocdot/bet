<x-filament-panels::page>
    @vite(['resources/css/app.css', 'resources/js/app.js'])
    <div class="space-y-6">
        {{ $this->form }}

        @if (! $agency)
            <div class="rounded-2xl border border-slate-200 bg-white p-6 text-sm text-slate-500 shadow-sm">
                Chọn một đại lý để xem dữ liệu 
            </div>
        @else
            <div class="grid gap-4 md:grid-cols-3">
                <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
                    <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Người chơi</p>
                    <p class="mt-2 text-3xl font-black text-slate-900">{{ number_format($summary['players']) }}</p>
                    <p class="mt-2 text-sm text-slate-500">Số người chơi thuộc tuyến đại lý đang chọn.</p>
                </div>
                <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
                    <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Giao dịch</p>
                    <p class="mt-2 text-3xl font-black text-slate-900">{{ number_format($summary['transactions']) }}</p>
                    <p class="mt-2 text-sm text-slate-500">Tổng giao dịch phát sinh từ tuyến này.</p>
                </div>
                <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
                    <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Yêu cầu rút</p>
                    <p class="mt-2 text-3xl font-black text-slate-900">{{ number_format($summary['withdrawals']) }}</p>
                    <p class="mt-2 text-sm text-slate-500">Tổng yêu cầu rút tiền từ tuyến này.</p>
                </div>
            </div>

            <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
                <div class="flex flex-wrap items-center justify-between gap-3">
                    <div>
                        <h3 class="text-lg font-black text-slate-900">Thông tin đại lý</h3>
                        <p class="text-sm text-slate-500">Đại lý hiện đang được chọn để theo dõi.</p>
                    </div>
                    <div class="rounded-full bg-rose-50 px-4 py-2 text-sm font-bold text-rose-700">
                        Mã giới thiệu: {{ $agency->affiliateProfile?->ref_code ?: 'Chưa có' }}
                    </div>
                </div>
                <div class="mt-4 grid gap-3 md:grid-cols-3">
                    <div class="rounded-xl bg-slate-50 px-4 py-3">
                        <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">ID đại lý</p>
                        <p class="mt-1 text-base font-black text-slate-900">{{ $agency->id }}</p>
                    </div>
                    <div class="rounded-xl bg-slate-50 px-4 py-3">
                        <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Tên đại lý</p>
                        <p class="mt-1 text-base font-black text-slate-900">{{ $agency->name }}</p>
                    </div>
                    <div class="rounded-xl bg-slate-50 px-4 py-3">
                        <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Số điện thoại</p>
                        <p class="mt-1 text-base font-black text-slate-900">{{ $agency->phone ?: '—' }}</p>
                    </div>
                </div>
            </div>

            <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
                <div class="mb-4">
                    <h3 class="text-lg font-black text-slate-900">Người chơi thuộc tuyến</h3>
                    <p class="text-sm text-slate-500">Danh sách người chơi do đại lý này quản lý.</p>
                </div>
                <div class="overflow-x-auto">
                    <table class="min-w-full divide-y divide-slate-200 text-sm">
                        <thead class="bg-slate-50">
                            <tr>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">ID</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Người chơi</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">SĐT</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Mã giới thiệu</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Tạo lúc</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-slate-100">
                            @forelse ($players as $player)
                                <tr>
                                    <td class="px-4 py-3 font-bold text-slate-900">{{ $player->id }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ $player->name }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ $player->phone ?: '—' }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ $player->affiliateProfile?->ref_code ?: '—' }}</td>
                                    <td class="px-4 py-3 text-slate-500">{{ $player->created_at?->format('d/m/Y H:i') }}</td>
                                </tr>
                            @empty
                                <tr>
                                    <td colspan="5" class="px-4 py-6 text-center text-slate-500">Chưa có người chơi thuộc tuyến đại lý này.</td>
                                </tr>
                            @endforelse
                        </tbody>
                    </table>
                </div>
                <div class="mt-4">
                    {{ $players->links() }}
                </div>
            </div>

            <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
                <div class="mb-4">
                    <h3 class="text-lg font-black text-slate-900">Giao dịch</h3>
                    <p class="text-sm text-slate-500">Các giao dịch phát sinh từ user thuộc tuyến đại lý.</p>
                </div>
                <div class="overflow-x-auto">
                    <table class="min-w-full divide-y divide-slate-200 text-sm">
                        <thead class="bg-slate-50">
                            <tr>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Mã GD</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Người chơi</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Loại</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Số tiền</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Trạng thái</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Tạo lúc</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-slate-100">
                            @forelse ($transactions as $transaction)
                                <tr>
                                    <td class="px-4 py-3 font-bold text-slate-900">{{ $transaction->client_ref ?: '#'.$transaction->id }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ $transaction->user?->name ?: '—' }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ $this->transactionTypeLabel((int) $transaction->type->value) }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ number_format((float) $transaction->amount, 0, ',', '.') }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ $this->transactionStatusLabel((int) $transaction->status->value) }}</td>
                                    <td class="px-4 py-3 text-slate-500">{{ $transaction->created_at?->format('d/m/Y H:i') }}</td>
                                </tr>
                            @empty
                                <tr>
                                    <td colspan="6" class="px-4 py-6 text-center text-slate-500">Chưa có giao dịch nào trong tuyến đại lý này.</td>
                                </tr>
                            @endforelse
                        </tbody>
                    </table>
                </div>
                <div class="mt-4">
                    {{ $transactions->links() }}
                </div>
            </div>

            <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
                <div class="mb-4">
                    <h3 class="text-lg font-black text-slate-900">Yêu cầu rút</h3>
                    <p class="text-sm text-slate-500">Các yêu cầu rút tiền của user thuộc tuyến đại lý.</p>
                </div>
                <div class="overflow-x-auto">
                    <table class="min-w-full divide-y divide-slate-200 text-sm">
                        <thead class="bg-slate-50">
                            <tr>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">ID</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Người chơi</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Số tiền</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Phí</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Thực nhận</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Trạng thái</th>
                                <th class="px-4 py-3 text-left font-bold text-slate-600">Tạo lúc</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-slate-100">
                            @forelse ($withdrawals as $withdrawal)
                                <tr>
                                    <td class="px-4 py-3 font-bold text-slate-900">#{{ $withdrawal->id }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ $withdrawal->user?->name ?: '—' }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ number_format((float) $withdrawal->amount, 0, ',', '.') }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ number_format((float) $withdrawal->fee, 0, ',', '.') }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ number_format((float) $withdrawal->net_amount, 0, ',', '.') }}</td>
                                    <td class="px-4 py-3 text-slate-700">{{ $this->withdrawalStatusLabel((int) $withdrawal->status->value) }}</td>
                                    <td class="px-4 py-3 text-slate-500">{{ $withdrawal->created_at?->format('d/m/Y H:i') }}</td>
                                </tr>
                            @empty
                                <tr>
                                    <td colspan="7" class="px-4 py-6 text-center text-slate-500">Chưa có yêu cầu rút nào trong tuyến đại lý này.</td>
                                </tr>
                            @endforelse
                        </tbody>
                    </table>
                </div>
                <div class="mt-4">
                    {{ $withdrawals->links() }}
                </div>
            </div>
        @endif
    </div>
</x-filament-panels::page>
