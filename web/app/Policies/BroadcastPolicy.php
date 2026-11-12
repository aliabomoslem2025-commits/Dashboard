<?php
namespace App\Policies;
use App\Models\User;
class BroadcastPolicy {
 public function viewAny(User $actor): bool { return $actor->hasAnyRole(['super_admin','admin','moderator']); }
 public function create(User $actor): bool { return $this->viewAny($actor); }
 public function view(User $actor,?int $id=null): bool { return $this->viewAny($actor); }
}
