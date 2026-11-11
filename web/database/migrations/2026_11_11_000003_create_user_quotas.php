<?php
use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;
return new class extends Migration { public function up(): void { Schema::create('user_quotas',function(Blueprint $t){$t->string('user_id')->primary();$t->unsignedBigInteger('max_total_bytes')->default(0);$t->unsignedBigInteger('max_file_bytes')->default(0);$t->string('period',16)->default('lifetime');$t->foreignId('updated_by')->nullable()->constrained('users')->nullOnDelete();$t->timestamps();});} public function down(): void {Schema::dropIfExists('user_quotas');} };
