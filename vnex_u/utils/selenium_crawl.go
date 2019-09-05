package utils

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	SU "../../selenium_utils"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/tebeka/selenium"
	"gopkg.in/mgo.v2/bson"
)

const (
	LinkRexp = `\b(https?)`
)

// get list of comment from box_comment
func GetAllDetailCmts(detailDriver selenium.WebDriver,
	linkCrwl LinkCrwl,
	vnexUsersC *mongo.Collection) []DetailComment {
	var allComments = make(map[string]DetailComment)

	commentItemE, er := detailDriver.FindElements(
		selenium.ByCSSSelector, ".comment_item")
	if er != nil {
		log.Println("eror on get comment_item value, ", er.Error(), linkCrwl.Link)
		return []DetailComment{}
	}

	// get text full comment
	// insert list of comment and user info
	for indexCmtItem := 0; indexCmtItem < len(commentItemE); indexCmtItem++ {
		// click txt_view_more comment
		countE := 0
		var cmtItem = commentItemE[indexCmtItem]
		for {
			viewMoreRepE, er := cmtItem.FindElements(
				selenium.ByCSSSelector, ".txt_view_more a.view_all_reply")
			fmt.Println("Error txt_view_more, ", er, len(viewMoreRepE))
			if len(viewMoreRepE) == 0 {
				break
			}

			for idxViewMore := 0; idxViewMore < len(viewMoreRepE); idxViewMore++ {
				// click view more
				efu := SU.NewEFU(detailDriver, 3)
				_, er = efu.WaitUntilClickable(viewMoreRepE[idxViewMore],
					".txt_view_more a.view_all_reply", -1)
				if er != nil {
					countE++
					break
				}
				href, _ := viewMoreRepE[idxViewMore].Text()
				fmt.Println("txt view more click err, ", len(viewMoreRepE), href, linkCrwl.Link)
				time.Sleep(1000 * time.Millisecond)

			}
			fmt.Println("-----------------------error count, ", len(viewMoreRepE))
			if countE > 0 {
				break
			}
		}

		// click icon_show_full_comment
		for {
			viewFullCmtsE, er := cmtItem.FindElements(
				selenium.ByCSSSelector,
				".content_less .icon_show_full_comment")
			fmt.Println("Error icon_show_full_comment, ", er, len(viewFullCmtsE))
			if len(viewFullCmtsE) == 0 {
				break
			}

			for idxViewFull := 0; idxViewFull < len(viewFullCmtsE); idxViewFull++ {
				// click view more
				efu := SU.NewEFU(detailDriver, 3)
				_, er = efu.WaitUntilClickable(viewFullCmtsE[idxViewFull],
					".txt_view_more a.view_all_reply", -1)
				if er != nil {
					countE++
					break
				}
				time.Sleep(2000 * time.Millisecond)
			}
			break
		}

		// check length sub_comment length class:
		// == 0 => ((full_content, user_status) | (content_more, user_status))
		// (> 0) => ((full_content, user_status) | (content_more, user_status)),
		// (sub_comment[sub_comment_item(full_content, user_status)])

		// check length .sub_comment
		subCmtE, _ := cmtItem.FindElements(
			selenium.ByCSSSelector, ".sub_comment_item")
		fmt.Println("len sub commenttttttttttttttttttttttttttt, ", len(subCmtE))
		if len(subCmtE) == 0 {
			detailCmt, er := GetDetailCmt(cmtItem, linkCrwl)
			if er != nil || detailCmt.Content == "" {
				if er != nil {
					log.Println(er.Error())
				} else {
					fmt.Println("content empty")
				}
				continue
			}
			// check detail comment not exist in map comments <= add
			if _, ok := allComments[detailCmt.Content]; !ok {
				allComments[detailCmt.Content] = detailCmt
			}
		} else {
			// get first comment
			firstComment, er := GetDetailCmt(cmtItem, linkCrwl)
			if er != nil || firstComment.Content == "" {
				if er != nil {
					log.Println(er.Error())
				} else {
					fmt.Println("content empty")
				}
				continue
			}
			// check detail comment not exist in map comments <= add
			if _, ok := allComments[firstComment.Content]; !ok {
				allComments[firstComment.Content] = firstComment
			}

			// get list of reply comment on sub_comment detail
			for i := 0; i < len(subCmtE); i++ {
				fmt.Println("reply comment, ", subCmtE[i])
				replyComment, er := GetDetailCmt(subCmtE[i], linkCrwl)
				if er != nil || replyComment.Content == "" {
					if er != nil {
						log.Println(er.Error())
					} else {
						fmt.Println("content empty")
					}
					continue
				}
				// check detail comment not exist in map comments <= add
				if _, ok := allComments[replyComment.Content]; !ok {
					allComments[replyComment.Content] = replyComment
				}
			}
		}
	}

	detailComments := []DetailComment{}
	for _, detail := range allComments {

		linkPattern := regexp.MustCompile(LinkRexp)
		// only insert user if can  match link profile pattern
		if linkPattern.MatchString(detail.ProfileLink) {
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			count, er := vnexUsersC.Count(ctx,
				bson.M{"profile_link": detail.ProfileLink})

			if er != nil {
				log.Println("Error, can not count profile from vnexpress_users, ", er.Error())
				continue
			}

			// only insert if not exist
			if count == int64(0) {
				user := VNExUser{
					ID:          primitive.NewObjectID(),
					ProfileLink: detail.ProfileLink,
					Status:      0}

				ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
				_, er := vnexUsersC.InsertOne(
					ctx,
					&user)
				if er != nil {
					log.Println("Error, can not insert user, ", user, er.Error())
				}
			}
		}
		detailComments = append(detailComments, detail)
	}
	return detailComments
}

// get detail comment {user_profile, comment}
func GetDetailCmt(cmtItem selenium.WebElement,
	linkCrwl LinkCrwl) (DetailComment, error) {

	detailComment := DetailComment{}
	// get user_info from comment element
	var userInfoE selenium.WebElement
	var er error

	// get user-info from .nickname
	efu := SU.NewEFU(nil, 10)
	userInfoE, er = efu.WaitElementWTimeOut(cmtItem, ".nickname", int64(2))
	if er != nil && er.Error() == SU.NoSuchElement {
		// get user-info from .avata_coment
		efu := SU.NewEFU(nil, 10)
		userInfoE, er = efu.WaitElementWTimeOut(cmtItem, ".avata_coment", int64(2))
		if er != nil {
			return detailComment, er
		}
	}

	profileLink, er := userInfoE.GetAttribute("href")
	if er != nil {
		log.Println("Error, can not get profile user, ", linkCrwl.Link, er.Error())
		return detailComment, er
	}

	userName, er := userInfoE.Text()
	if er != nil {
		log.Println("Error, can not get user name, ", linkCrwl.Link, er.Error())
		return detailComment, er
	}

	detailComment.ProfileLink = profileLink
	detailComment.UserName = userName

	// click view full_content comment
	var fullCmtText = ""
	// click view content_more
	efu = SU.NewEFU(nil, 10)
	cmtsE, fullCmter := efu.WaitElementWTimeOut(cmtItem, "p.full_content", int64(2))
	if fullCmter == nil {
		fullCmtText, er = cmtsE.Text()
		if er != nil {
			return detailComment, er
		}
	}
	if fullCmter != nil {
		// click view content_more
		efu := SU.NewEFU(nil, 10)
		cmtmoresE, er := efu.WaitElementWTimeOut(cmtItem, ".content_more", int64(2))

		if er != nil {
			return detailComment, er
		}

		fullCmtText, er = cmtmoresE.Text()
		if er != nil {
			return detailComment, er
		}
	}
	if fullCmtText == "" {
		return detailComment, er
	}
	detailComment.Content = fullCmtText
	return detailComment, nil
}
