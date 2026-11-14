<?php
use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration {
    public function up(): void
    {
        Schema::table('users', function (Blueprint $table) {
            $table->string('username', 64)->nullable()->unique();
        });
        DB::table('users')->whereNull('username')->orderBy('id')->chunkById(100, function ($users) {
            foreach ($users as $user) {
                $base = preg_replace('/[^A-Za-z0-9._-]/', '', strstr((string) $user->email, '@', true) ?: 'admin');
                $base = substr($base ?: 'admin', 0, 48);
                $candidate = $base; $suffix = 1;
                while (DB::table('users')->where('username', $candidate)->exists()) {
                    $candidate = substr($base, 0, 48).$suffix++;
                }
                DB::table('users')->where('id', $user->id)->update(['username' => $candidate]);
            }
        });
        Schema::table('users', function (Blueprint $table) {
            $table->string('username', 64)->nullable(false)->change();
        });
    }

    public function down(): void
    {
        Schema::table('users', function (Blueprint $table) { $table->dropUnique(['username']); $table->dropColumn('username'); });
    }
};
