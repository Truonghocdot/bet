# Cloudflare R2 backups

The backup service stores completed, private snapshots under
`BACKUP_PREFIX/BACKUP_SITE/YYYY/MM/DD/`, with completion manifests in a small
`_manifests/` catalog. Each snapshot contains gzipped JSONL
exports for users, wallets, supporting transactions, the full affiliate graph,
news articles, and banners. Original files under `storage/app/public/news` and
`storage/app/public/banners` are included. The manifest is uploaded last and
contains a SHA-256 checksum for every object.

New snapshots use `site-backup-v2`. Existing `site-backup-v1` snapshots remain
available to list, verify, and restore with their original data scope.

## R2 setup

Create a private R2 bucket and an R2 S3 API token restricted to Object Read &
Write for that bucket. Configure a different `BACKUP_SITE` for each deployment,
even when both deployments share one bucket.

Cloudflare references: [create R2 API tokens](https://developers.cloudflare.com/r2/api/tokens/)
and [S3 API compatibility](https://developers.cloudflare.com/r2/api/s3/api/).

```dotenv
BACKUP_DISK=r2
BACKUP_SITE=practice
BACKUP_PREFIX=backups
BACKUP_RETENTION_DAYS=30
BACKUP_KEEP_MINIMUM=7
BACKUP_SCHEDULE=02:15

R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET=
R2_ENDPOINT=https://<ACCOUNT_ID>.r2.cloudflarestorage.com
```

Run one backup and verify it before enabling the scheduler:

```bash
php artisan backup:create
php artisan backup:list
php artisan backup:verify 'YYYY/MM/DD/HHMMSS-UUID'
```

Laravel's scheduler must be invoked once per minute by cron:

```cron
* * * * * cd /path/to/admin && php artisan schedule:run >> /dev/null 2>&1
```

The scheduled task creates one snapshot per day and prunes expired snapshots.
It always retains at least `BACKUP_KEEP_MINIMUM` completed snapshots.

## Restore

Verify and restore a snapshot with:

```bash
php artisan backup:restore 'YYYY/MM/DD/HHMMSS-UUID'
```

Restore uses upserts so matching rows and images are restored without deleting
data created after the snapshot. Run migrations first when restoring into an
empty database. Use separate credentials with read-only bucket access on a
restore host whenever possible.
