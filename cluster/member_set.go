package cluster

// MemberSet represents a set of members in the cluster.
type MemberSet struct {
	members map[string]*Member
}

// NewMemberSet creates a new MemberSet with the given members.
func NewMemberSet(members ...*Member) *MemberSet {
	m := make(map[string]*Member)
	for _, member := range members {
		m[member.ID] = member
	}
	return &MemberSet{
		members: m,
	}
}

// Len returns the number of members in the set.
func (s *MemberSet) Len() int {
	return len(s.members)
}

// GetByHost returns the member with the given host, or nil if not found.
func (s *MemberSet) GetByHost(host string) *Member {
	for _, member := range s.members {
		if member.Host == host {
			return member
		}
	}
	return nil
}

// Add adds a member to the set.
func (s *MemberSet) Add(member *Member) {
	s.members[member.ID] = member
}

// Contains checks if the set contains the given member.
func (s *MemberSet) Contains(member *Member) bool {
	_, ok := s.members[member.ID]
	return ok
}

// Remove removes a member from the set.
func (s *MemberSet) Remove(member *Member) {
	delete(s.members, member.ID)
}

// RemoveByHost removes a member with the given host from the set.
func (s *MemberSet) RemoveByHost(host string) {
	member := s.GetByHost(host)
	if member != nil {
		s.Remove(member)
	}
}

// Slice returns a slice of all members in the set.
func (s *MemberSet) Slice() []*Member {
	members := make([]*Member, 0, len(s.members))
	for _, member := range s.members {
		members = append(members, member)
	}
	return members
}

// ForEach iterates over all members in the set and applies the given function.
func (s *MemberSet) ForEach(fun func(m *Member) bool) {
	for _, member := range s.members {
		if !fun(member) {
			break
		}
	}
}

// Except returns a slice of members that are not in the given slice.
func (s *MemberSet) Except(members []*Member) []*Member {
	except := []*Member{}
	m := make(map[string]*Member)
	for _, member := range members {
		m[member.ID] = member
	}
	for _, member := range s.members {
		if _, ok := m[member.ID]; !ok {
			except = append(except, member)
		}
	}
	return except
}

// FilterByKind returns a slice of members that have the given kind.
func (s *MemberSet) FilterByKind(kind string) []*Member {
	members := []*Member{}
	for _, member := range s.members {
		if member.HasKind(kind) {
			members = append(members, member)
		}
	}
	return members
}
