<?php
namespace Database\Seeders;
use App\Models\User;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\DB;
class DatabaseSeeder extends Seeder { public function run(): void { foreach ([['super_admin','مدیر ارشد'],['admin','مدیر'],['moderator','ناظر'],['viewer','مشاهده‌گر']] as [$name,$label]) DB::table('roles')->updateOrInsert(['name'=>$name],['label'=>$label,'updated_at'=>now(),'created_at'=>now()]);$email=env('INITIAL_ADMIN_EMAIL');$password=env('INITIAL_ADMIN_PASSWORD');if($email&&$password){$user=User::firstOrCreate(['email'=>$email],['name'=>'مدیر ارشد','password'=>$password,'is_active'=>true]);$user->roles()->syncWithoutDetaching([DB::table('roles')->where('name','super_admin')->value('id')]);}} }
