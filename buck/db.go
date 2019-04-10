package buck

import (
	"github.com/strongjz/slack-bucks/database"
	"log"
)

func (b *Buck) updateDB(g database.Gift) error {

	log.Print("[INFO] updateDB")

	err := b.db.WriteGift(&g)
	if err != nil {

		return err
	}

	return nil
}
