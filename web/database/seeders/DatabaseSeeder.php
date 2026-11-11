<?php
namespace Database\Seeders;
use App\Models\User;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\DB;
class DatabaseSeeder extends Seeder {
 public function run(): void {
  foreach ([['super_admin','مدیر ارشد'],['admin','مدیر'],['moderator','ناظر'],['viewer','مشاهده‌گر']] as [$name,$label]) DB::table('roles')->updateOrInsert(['name'=>$name],['label'=>$label,'updated_at'=>now(),'created_at'=>now()]);
  DB::table('activity_thresholds')->insertOrIgnore(['id'=>1,'heavy_min_messages'=>1000,'active_min_messages'=>100,'low_active_min_messages'=>10,'dormant_days'=>30,'suspicious_new_account_days'=>1,'suspicious_messages_per_hour'=>5000,'created_at'=>now(),'updated_at'=>now()]);
  foreach(['cpu_percent'=>85,'ram_percent'=>85,'disk_percent'=>90] as $name=>$value) DB::table('metric_thresholds')->insertOrIgnore(['name'=>$name,'value'=>$value,'created_at'=>now(),'updated_at'=>now()]);
  $email=env('INITIAL_ADMIN_EMAIL');$password=env('INITIAL_ADMIN_PASSWORD');
  if($email&&$password){$user=User::firstOrCreate(['email'=>$email],['name'=>'مدیر ارشد','password'=>$password,'is_active'=>true]);$user->roles()->syncWithoutDetaching([DB::table('roles')->where('name','super_admin')->value('id')]);}
 }
}
