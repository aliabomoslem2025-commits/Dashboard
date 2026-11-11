<?php
namespace App\Policies;
use App\Models\User;
class SynapseRoomPolicy {
 public function viewAny(User $actor): bool { return $actor->hasAnyRole(['super_admin','admin','moderator','viewer']); }
 public function view(User $actor,string $roomID): bool { return $this->viewAny($actor); }
 public function update(User $actor,string $roomID): bool { return $actor->hasAnyRole(['super_admin','admin']); }
 public function moderate(User $actor,string $roomID): bool { return $actor->hasAnyRole(['super_admin','admin','moderator']); }
 public function delete(User $actor,string $roomID): bool { return $actor->hasAnyRole(['super_admin','admin']); }
}
