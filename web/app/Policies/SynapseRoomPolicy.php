<?php
namespace App\Policies;
use App\Models\User;
class SynapseRoomPolicy {
 public function viewAny(User $actor): bool { return $actor->hasAnyRole(['super_admin','admin','moderator','viewer']); }
 public function view(User $actor): bool { return $this->viewAny($actor); }
 public function moderate(User $actor): bool { return $actor->hasAnyRole(['super_admin','admin','moderator']); }
 public function delete(User $actor): bool { return $actor->hasRole('super_admin'); }
}
