<?php
namespace App\Policies;
use App\Models\User;
class SynapseUserPolicy {
 public function viewAny(User $actor): bool { return $actor->hasAnyRole(['super_admin','admin','moderator','viewer']); }
 public function view(User $actor,string $userID): bool { return $this->viewAny($actor); }
 public function create(User $actor): bool { return $actor->hasAnyRole(['super_admin','admin']); }
 public function update(User $actor,string $userID): bool { return $actor->hasAnyRole(['super_admin','admin']); }
 public function moderate(User $actor,string $userID): bool { return $actor->hasAnyRole(['super_admin','admin','moderator']); }
 public function delete(User $actor,string $userID): bool { return $actor->hasRole('super_admin'); }
}
