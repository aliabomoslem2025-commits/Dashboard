<?php
namespace Database\Seeders;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\DB;
class DatabaseSeeder extends Seeder { public function run(): void { foreach ([['super_admin','مدیر ارشد'],['admin','مدیر'],['moderator','ناظر'],['viewer','مشاهده‌گر']] as [$name,$label]) DB::table('roles')->updateOrInsert(['name'=>$name],['label'=>$label,'updated_at'=>now(),'created_at'=>now()]); } }
